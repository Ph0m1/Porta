package proxy

import (
	"context"
	"net/url"
	"time"

	"github.com/ph0m1/p_gateway/config"
	"github.com/ph0m1/p_gateway/sd"
)

func NewRoundRobinLoadBalancedMiddleware(remote *config.Backend) Middleware {
	return newLoadBalancedMiddleware(sd.NewRoundRobinLB(sd.FixedSubscriber(remote.Host)))
}

func NewRandomLoadBalancedMiddleware(remote *config.Backend) Middleware {
	return newLoadBalancedMiddleware(sd.NewRandomLB(sd.FixedSubscriber(remote.Host), time.Now().UnixNano()))
}

func newLoadBalancedMiddleware(lb sd.Balancer) Middleware {
	return func(next ...Proxy) Proxy {
		if len(next) > 1 {
			panic(ErrTooManyProxies)
		}
		return func(ctx context.Context, request *Request) (*Response, error) {
			host, err := lb.Host()
			if err != nil {
				return nil, err
			}
			r := request.Clone()

			rawURL := []byte{}
			rawURL = append(rawURL, host...)
			rawURL = append(rawURL, r.Path...)
			r.URL, err = url.Parse(string(rawURL))
			if err != nil {
				return nil, err
			}
			r.URL.RawQuery = r.Query.Encode()

			return next[0](ctx, &r)
		}
	}
}
