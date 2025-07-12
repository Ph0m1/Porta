package proxy

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/ph0m1/porta/config"
	"github.com/ph0m1/porta/encoding"
)

// ErrInvalidStatusCode is the error returned by the http proxy when the
// received status code of the response is not 200 or 201
var ErrInvalidStatusCode = errors.New("Invalid status code")

// creates http client based with the received context
type HTTPClientFactory func(ctx context.Context) *http.Client

func NewHttpClient(_ context.Context) *http.Client {
	// 创建一个不使用代理的 HTTP 客户端
	client := &http.Client{
		Transport: &http.Transport{
			Proxy: nil, // 禁用代理
		},
	}
	return client
}

func httpProxy(backend *config.Backend) Proxy {
	return NewHttpProxy(backend, NewHttpClient, backend.Decoder)
}

func NewRequestBuilderMiddleware(remote *config.Backend) Middleware {
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
		// 添加调试信息
		fmt.Printf("[DEBUG] Backend response status: %d\n", resp.StatusCode)
		fmt.Printf("[DEBUG] Backend response headers: %v\n", resp.Header)

		if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
			fmt.Printf("[DEBUG] Invalid status code: %d\n", resp.StatusCode)
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
