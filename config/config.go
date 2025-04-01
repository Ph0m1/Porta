package config

import (
	"fmt"
	"github.com/ph0m1/p_gateway/encoding"
	"regexp"
	"time"
)

// ServiceConfig defines the service
type ServiceConfig struct {
	// set of endpoint definitions
	Endpoints []*EndpointConfig `mapstructure:"endpoints"`
	// default timeout
	Timeout time.Duration `mapstructure:"timeout"`
	// default TTL for GET
	CacheTTL time.Duration `mapstructure:"cache_ttl"`
	// default set of hosts
	Host []string `mapstructure:"host"`
	// port to bind service
	Port int `mapstructure:"port"`
	// version code of the configuration
	Version int `mapstructure:"version"`

	// run in Debug Mode
	Debug bool
}

// EndpointConfig defines the configuration of a single endpoint to be exposed by service
type EndpointConfig struct {
	// url pattern to be registered and exposed to the world
	Endpoint string `mapstructure:"endpoint"`
	// HTTP method of the endpoint (GET, POST, PUT, etc)
	Method string `mapstructure:"method"`
	// set of definitions of the backends to be linked to this endpoint
	Backend []Backend `mapstructure:"backend"`
	// number of concurrent calls this endpoint must send to the backends
	ConcurrentCalls int `mapstructure:"concurrent_calls"`
	// timeout of this endpoint
	Timeout time.Duration `mapstructure:"timeout"`
	// duration of cache header
	CacheTTL time.Duration `mapstructure:"cache_ttl"`
	// list of query string params to be extracted from the URI
	QueryString []string `mapstructure:"querystring_params"`
}

// Backend defines how to connect to the backend service and how to process the received response
type Backend struct {
	// the name of the group the response should be moved to
	Group string `mapstructure:"group"`
	// HTTP method of the request to send to the backend
	Method string `mapstructure:"method"`
	// Set of hosts of the API
	Host []string `mapstructure:"host"`
	// URL pattern to use to locate the resource to be consumed
	URLPattern string `mapstructure:"url_pattern"`
	// set of response fields to remove
	Blacklist []string `mapstructure:"blacklist"`
	// set of response fields to allow
	Whitelist []string `mapstructure:"whitelist"`
	// map of response fields to renamed and their new names
	Mapping map[string]string `mapstructure:"mapping"`
	// the encoding format
	Encoding string `mapstructure:"encoding"`
	// name of the field to extract to the root
	Target string `mapstructure:"target"`

	// list of keys to be replaced in the URLPattern
	URLKeys []string
	// number of concurrent calls this endpoint must send to the API
	ConcurrentCalls int
	// timeout of this backend
	Timeout time.Duration
	// decoder to use in order to parse the received response from the API
	Decoder encoding.Decoder
}

var (
	hostPattern = regexp.MustCompile(`(https?://)?([a-zA-Z\-_0-9]+)(:[0-9]{2,6})?/?`)
	defaultPort = 8080
)

func (s *ServiceConfig) Init() error {
	if s.Version != 1 {
		return fmt.Errorf("Unsupported version: %d\n", s.Version)
	}
	if s.Port == 0 {
		s.Port = defaultPort
	}
	s.Host = s.cleanHost(s.Host)
}

func (s *ServiceConfig) cleanHost(host string) string {

}
