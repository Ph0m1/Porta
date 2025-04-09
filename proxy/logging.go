package proxy

import (
	"context"
	"time"

	"github.com/ph0m1/p_gateway/logging"
)

func NewLoggingMiddleware(logger logging.Logger, name string) Middleware {
	return func(next ...Proxy) Proxy {
		if len(next) > 1 {
			panic(ErrTooManyProxies)
		}
		return func(ctx context.Context, request *Request) (*Response, error) {
			begin := time.Now()
			logger.Info(name, "Calling backend")
			logger.Debug("Request", request)

			result, err := next[0](ctx, request)

			logger.Info(name, "Call to backend took", time.Since(begin).String())
			if err != nil {
				logger.Warning(name, "Call to backend failed:", err.Error())
			}
			return result, err
		}
	}
}
