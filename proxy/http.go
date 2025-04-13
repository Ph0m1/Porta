package proxy

import (
	"context"
	"errors"
	"net/http"

	"github.com/ph0m1/p_gateway/config"
	"github.com/ph0m1/p_gateway/encoding"
)

var ErrInvalidStatusCode = errors.New("Invalid status code")

// creates http client based with the received context
type HTTPClientFactory func(ctx context.Context) *http.Client

func NewHttpClient(_ context.Context) *http.Client { return http.DefaultClient }

func httpProxy(backend *config.Backend) Proxy {
	return NewHttpProxy(backend, NewHttpClient, backend.Decoder)
}

func NewRequestBuilderMiuddleware(remote *config.Backend) Middleware {
	return func(next ...Proxy) Proxy {
		if len(next) > 1 {
			panic(ErrTooManyProxies)
		}
		return func(ctx context.Context, request *Request) (*Response, error) {
			r := request.Clone()
			r.GeneratePath(remote.URLPattern)
			r.Method = remote.Method
			return next[0](ctx, &r)
		}
	}
}

func NewHttpProxy(remote *config.Backend, clientFactory HTTPClientFactory, decode encoding.Decoder) Proxy {
	formatter := NewEntityFormatter(remote.Target, remote.Whitelist, remote.Blacklist, remote.Group, remote.Mapping)

	return func(ctx context.Context, request *Request) (*Response, error) {
		requestToBackend, err := http.NewRequest(request.Method, request.URL.String(), request.Body)
		if err != nil {
			return nil, err
		}
		requestToBackend.Header = request.Headers

		resp, err := clientFactory(ctx).Do(requestToBackend.WithContext(ctx))
		requestToBackend.Body.Close()
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:

		}
		if err != nil {
			return nil, err
		}
		if resp.StatusCode != http.StatusCreated {
			return nil, ErrInvalidStatusCode
		}
		var data map[string]interface{}
		err = decode(resp.Body, &data)
		resp.Body.Close()
		if err != nil {
			return nil, err
		}
		r := formatter.Format(Response{data, true})
		return &r, nil
	}
}
