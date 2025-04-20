package gin

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/ph0m1/p_gateway/config"
	"github.com/ph0m1/p_gateway/proxy"
)

var ErrInternalError = errors.New("internal server error")

type HandlerFactory func(endpointConfig *config.EndpointConfig, proxy2 proxy.Proxy) gin.HandlerFunc

func EndpointHandler(cfg *config.EndpointConfig, proxy proxy.Proxy) gin.HandlerFunc {
	endpointTimeout := time.Duration(cfg.Timeout) * time.Millisecond

	return func(c *gin.Context) {
		requestCtx, cancel := context.WithTimeout(c, endpointTimeout)

		c.Header("X_X", "Version undefined")

		response, err := proxy(requestCtx, NewRequest(c, cfg.QueryString))
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			cancel()
			return
		}

		select {
		case <-requestCtx.Done():
			c.AbortWithError(http.StatusInternalServerError, ErrInternalError)
			cancel()
		default:
		}

		if cfg.CacheTTL.Seconds() != 0 && response != nil && response.IsComplete {
			c.Header("Cache-Control", fmt.Sprintf("public, max-age=%d", int(cfg.CacheTTL.Seconds())))
			c.JSON(http.StatusOK, response.Data)
			cancel()
			return
		}
		c.JSON(http.StatusOK, gin.H{})
		cancel()
	}
}

var (
	headersToSend        = []string{"Content-Type"}
	userAgentHeaderValue = []string{"X_X Version undefined"}
)

func NewRequest(c *gin.Context, queryString []string) *proxy.Request {
	params := make(map[string]string, len(c.Params))
	for _, param := range c.Params {
		params[strings.Title(param.Key)] = param.Value
	}

	headers := make(map[string][]string, 2+len(headersToSend))
	headers["X-Forwarded-For"] = []string{c.ClientIP()}
	headers["User-Agent"] = userAgentHeaderValue
	for _, k := range headersToSend {
		if h, ok := c.Request.Header[k]; ok {
			headers[k] = h
		}
	}

	query := make(map[string][]string, len(queryString))
	for i := range queryString {
		if v := c.Request.URL.Query().Get(queryString[i]); v != "" {
			query[queryString[i]] = []string{v}
		}
	}

	return &proxy.Request{
		Method:  c.Request.Method,
		Query:   query,
		Body:    c.Request.Body,
		Params:  params,
		Headers: headers,
	}
}
