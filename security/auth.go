package security

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// AuthConfig holds authentication configuration
type AuthConfig struct {
	JWTSecret     string              `json:"jwt_secret"`
	JWTExpiration time.Duration       `json:"jwt_expiration"`
	APIKeys       map[string]string   `json:"api_keys"`       // key -> client_id
	BasicAuth     map[string]string   `json:"basic_auth"`     // username -> password
	RequiredRoles map[string][]string `json:"required_roles"` // endpoint -> roles
}

// Claims represents JWT claims
type Claims struct {
	UserID   string   `json:"user_id"`
	ClientID string   `json:"client_id"`
	Roles    []string `json:"roles"`
	jwt.RegisteredClaims
}

// AuthContext holds authentication information
type AuthContext struct {
	UserID     string
	ClientID   string
	Roles      []string
	AuthMethod string
}

// AuthMiddleware provides authentication middleware
type AuthMiddleware struct {
	config *AuthConfig
}

// NewAuthMiddleware creates a new authentication middleware
func NewAuthMiddleware(config *AuthConfig) *AuthMiddleware {
	return &AuthMiddleware{
		config: config,
	}
}

// Authenticate validates the request and returns auth context
func (am *AuthMiddleware) Authenticate(r *http.Request) (*AuthContext, error) {
	// Try JWT authentication first
	if authHeader := r.Header.Get("Authorization"); authHeader != "" {
		if strings.HasPrefix(authHeader, "Bearer ") {
			token := strings.TrimPrefix(authHeader, "Bearer ")
			return am.validateJWT(token)
		}

		if strings.HasPrefix(authHeader, "Basic ") {
			return am.validateBasicAuth(authHeader)
		}
	}

	// Try API Key authentication
	if apiKey := r.Header.Get("X-API-Key"); apiKey != "" {
		return am.validateAPIKey(apiKey)
	}

	// Try query parameter API key
	if apiKey := r.URL.Query().Get("api_key"); apiKey != "" {
		return am.validateAPIKey(apiKey)
	}

	return nil, errors.New("no valid authentication provided")
}

// validateJWT validates a JWT token
func (am *AuthMiddleware) validateJWT(tokenString string) (*AuthContext, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(am.config.JWTSecret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("invalid JWT token: %w", err)
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return &AuthContext{
			UserID:     claims.UserID,
			ClientID:   claims.ClientID,
			Roles:      claims.Roles,
			AuthMethod: "jwt",
		}, nil
	}

	return nil, errors.New("invalid JWT claims")
}

// validateAPIKey validates an API key
func (am *AuthMiddleware) validateAPIKey(apiKey string) (*AuthContext, error) {
	if clientID, exists := am.config.APIKeys[apiKey]; exists {
		return &AuthContext{
			ClientID:   clientID,
			Roles:      []string{"api_user"},
			AuthMethod: "api_key",
		}, nil
	}

	return nil, errors.New("invalid API key")
}

// validateBasicAuth validates basic authentication
func (am *AuthMiddleware) validateBasicAuth(authHeader string) (*AuthContext, error) {
	encoded := strings.TrimPrefix(authHeader, "Basic ")
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, errors.New("invalid basic auth encoding")
	}

	credentials := strings.SplitN(string(decoded), ":", 2)
	if len(credentials) != 2 {
		return nil, errors.New("invalid basic auth format")
	}

	username, password := credentials[0], credentials[1]
	if storedPassword, exists := am.config.BasicAuth[username]; exists && storedPassword == password {
		return &AuthContext{
			UserID:     username,
			Roles:      []string{"basic_user"},
			AuthMethod: "basic",
		}, nil
	}

	return nil, errors.New("invalid basic auth credentials")
}

// Authorize checks if the auth context has required permissions
func (am *AuthMiddleware) Authorize(authCtx *AuthContext, endpoint string) error {
	requiredRoles, exists := am.config.RequiredRoles[endpoint]
	if !exists {
		// No specific roles required for this endpoint
		return nil
	}

	// Check if user has any of the required roles
	for _, userRole := range authCtx.Roles {
		for _, requiredRole := range requiredRoles {
			if userRole == requiredRole {
				return nil
			}
		}
	}

	return fmt.Errorf("insufficient permissions: requires one of %v, has %v", requiredRoles, authCtx.Roles)
}

// HTTPMiddleware returns an HTTP middleware function
func (am *AuthMiddleware) HTTPMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip authentication for health checks and metrics
		if strings.HasPrefix(r.URL.Path, "/__health") ||
			strings.HasPrefix(r.URL.Path, "/__ready") ||
			strings.HasPrefix(r.URL.Path, "/__live") ||
			strings.HasPrefix(r.URL.Path, "/metrics") {
			next.ServeHTTP(w, r)
			return
		}

		authCtx, err := am.Authenticate(r)
		if err != nil {
			http.Error(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
			return
		}

		// Check authorization
		if err := am.Authorize(authCtx, r.URL.Path); err != nil {
			http.Error(w, "Forbidden: "+err.Error(), http.StatusForbidden)
			return
		}

		// Add auth context to request context
		ctx := context.WithValue(r.Context(), "auth", authCtx)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GenerateJWT generates a JWT token for a user
func (am *AuthMiddleware) GenerateJWT(userID, clientID string, roles []string) (string, error) {
	claims := &Claims{
		UserID:   userID,
		ClientID: clientID,
		Roles:    roles,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(am.config.JWTExpiration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "porta-gateway",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(am.config.JWTSecret))
}

// GetAuthContext extracts auth context from request context
func GetAuthContext(r *http.Request) (*AuthContext, bool) {
	authCtx, ok := r.Context().Value("auth").(*AuthContext)
	return authCtx, ok
}

// SignatureAuth provides request signature authentication
type SignatureAuth struct {
	secrets map[string]string // client_id -> secret
}

// NewSignatureAuth creates a new signature authentication
func NewSignatureAuth(secrets map[string]string) *SignatureAuth {
	return &SignatureAuth{
		secrets: secrets,
	}
}

// ValidateSignature validates request signature
func (sa *SignatureAuth) ValidateSignature(r *http.Request) (*AuthContext, error) {
	clientID := r.Header.Get("X-Client-ID")
	signature := r.Header.Get("X-Signature")
	timestamp := r.Header.Get("X-Timestamp")

	if clientID == "" || signature == "" || timestamp == "" {
		return nil, errors.New("missing signature headers")
	}

	secret, exists := sa.secrets[clientID]
	if !exists {
		return nil, errors.New("unknown client ID")
	}

	// Verify timestamp (prevent replay attacks)
	reqTime, err := time.Parse(time.RFC3339, timestamp)
	if err != nil {
		return nil, errors.New("invalid timestamp format")
	}

	if time.Since(reqTime) > 5*time.Minute {
		return nil, errors.New("request timestamp too old")
	}

	// Create signature string
	method := r.Method
	path := r.URL.Path
	query := r.URL.RawQuery

	signatureString := fmt.Sprintf("%s\n%s\n%s\n%s\n%s",
		method, path, query, timestamp, clientID)

	// Calculate expected signature
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(signatureString))
	expectedSignature := base64.StdEncoding.EncodeToString(h.Sum(nil))

	if !hmac.Equal([]byte(signature), []byte(expectedSignature)) {
		return nil, errors.New("invalid signature")
	}

	return &AuthContext{
		ClientID:   clientID,
		Roles:      []string{"signed_user"},
		AuthMethod: "signature",
	}, nil
}

// OAuth2Config holds OAuth2 configuration
type OAuth2Config struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	RedirectURL  string `json:"redirect_url"`
	AuthURL      string `json:"auth_url"`
	TokenURL     string `json:"token_url"`
	UserInfoURL  string `json:"user_info_url"`
}

// OAuth2Handler handles OAuth2 authentication flow
type OAuth2Handler struct {
	config *OAuth2Config
}

// NewOAuth2Handler creates a new OAuth2 handler
func NewOAuth2Handler(config *OAuth2Config) *OAuth2Handler {
	return &OAuth2Handler{
		config: config,
	}
}

// AuthURL generates OAuth2 authorization URL
func (oh *OAuth2Handler) AuthURL(state string) string {
	return fmt.Sprintf("%s?client_id=%s&redirect_uri=%s&response_type=code&state=%s",
		oh.config.AuthURL, oh.config.ClientID, oh.config.RedirectURL, state)
}

// ExchangeCode exchanges authorization code for access token
func (oh *OAuth2Handler) ExchangeCode(code string) (string, error) {
	// Implementation would make HTTP request to token endpoint
	// This is a simplified version
	return "access_token_placeholder", nil
}

// GetUserInfo retrieves user information using access token
func (oh *OAuth2Handler) GetUserInfo(accessToken string) (map[string]interface{}, error) {
	// Implementation would make HTTP request to user info endpoint
	// This is a simplified version
	return map[string]interface{}{
		"user_id": "oauth_user",
		"email":   "user@example.com",
		"roles":   []string{"oauth_user"},
	}, nil
}
