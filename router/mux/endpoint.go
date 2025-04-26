package mux

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ph0m1/p_gateway/config"
	"github.com/ph0m1/p_gateway/proxy"
	"net/http"
	"time"
)

var ErrInternalError = errors.New("internal server error")

type HandlerFactory func(*config.EndpointConfig, proxy.Proxy) http.HandlerFunc

func EndpointHandler(cfg *config.EndpointConfig, proxy proxy.Proxy) http.HandlerFunc {
	endpointTimeout := time.Duration(cfg.Timeout) * time.Millisecond

	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != cfg.Method {
			http.Error(w, "", http.StatusMethodNotAllowed)
			return
		}

		requestCtx, cancel := context.WithTimeout(context.Background(), endpointTimeout)

		w.Header().Set("X_X", "Version undefined")

		response, err := proxy(requestCtx, NewRequest(r, cfg.QueryString))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			cancel()
			return
		}

		select {
		case <-requestCtx.Done():
			http.Error(w, ErrInternalError.Error(), http.StatusInternalServerError)
			cancel()
			return
		default:
		}

		var js []byte

		if response != nil {
			js, err = json.Marshal(response.Data)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				cancel()
				return
			}
			if cfg.CacheTTL.Seconds() != 0 && response.IsComplete {
				w.Header().Set("Cache-Control", fmt.Sprintf("public, max-age=%d", int(cfg.CacheTTL.Seconds())))
			}
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(js)
		cancel()
	}
}

var (
	headersToSend        = []string{"Content-Type"}
	userAgentHeaderValue = []string{"X_X Version undefined"}
)

func NewRequest(r *http.Request, queryString []string) *proxy.Request {
	// params := make(map[string]string, len(c.Params))
	// for _, param := range c.Params {
	// 	params[strings.Title(param.Key)] = param.Value
	// }

	headers := make(map[string][]string, 2+len(headersToSend))
	headers["X-Forwarded-For"] = []string{r.RemoteAddr}
	headers["User-Agent"] = userAgentHeaderValue

	for _, k := range headersToSend {
		if h, ok := r.Header[k]; ok {
			headers[k] = h
		}
	}

	query := make(map[string][]string, len(queryString))
	for i := range queryString {
		if v := r.URL.Query().Get(queryString[i]); v != "" {
			query[queryString[i]] = []string{v}
		}
	}

	return &proxy.Request{
		Method:  r.Method,
		Query:   query,
		Body:    r.Body,
		Params:  map[string]string{},
		Headers: headers,
	}
}
