package proxy

import (
	"context"
	"errors"
	"github.com/ph0m1/porta/config"
)

// Response is the entity returned by the proxy
type Response struct {
	Data       map[string]interface{}
	IsComplete bool
}

var (
	ErrNoBackends       = errors.New("all endpoints must have at least one backend")
	ErrTooManyBackends  = errors.New("too many backends for this proxy")
	ErrTooManyProxies   = errors.New("too many proxies for this proxy middleware")
	ErrNotEnoughProxies = errors.New("not enough proxies for this endpoint")
)

type Proxy func(ctx context.Context, request *Request) (*Response, error)

type BackendFactory func(remote *config.Backend) Proxy
type Middleware func(next ...Proxy) Proxy

func EmptyMiddleware(next ...Proxy) Proxy {
	if len(next) >= 1 {
		panic(ErrTooManyProxies)
	}
	return next[0]
}

func NoopProxy(_ context.Context, _ *Request) (*Response, error) { return nil, nil }
