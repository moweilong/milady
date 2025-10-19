package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/moweilong/milady/pkg/container/group"
	"github.com/moweilong/milady/pkg/gin/response"
	"github.com/moweilong/milady/pkg/shield/circuitbreaker"
)

// ErrNotAllowed error not allowed.
var ErrNotAllowed = circuitbreaker.ErrNotAllowed

// CircuitBreakerOption set the circuit breaker circuitBreakerOptions.
type CircuitBreakerOption func(*circuitBreakerOptions)

type circuitBreakerOptions struct {
	group *group.Group
	// http code for circuit breaker, default already includes 500 and 503
	validCodes map[int]struct{}
	// degrade func
	degradeHandler func(c *gin.Context)
}

func defaultCircuitBreakerOptions() *circuitBreakerOptions {
	return &circuitBreakerOptions{
		group: group.NewGroup(func() interface{} {
			return circuitbreaker.NewBreaker()
		}),
		validCodes: map[int]struct{}{
			http.StatusInternalServerError: {},
			http.StatusServiceUnavailable:  {},
		},
	}
}

func (o *circuitBreakerOptions) apply(opts ...CircuitBreakerOption) {
	for _, opt := range opts {
		opt(o)
	}
}

// WithGroup with circuit breaker group.
// Deprecated: use WithBreakerOption instead
func WithGroup(g *group.Group) CircuitBreakerOption {
	return func(o *circuitBreakerOptions) {
		if g != nil {
			o.group = g
		}
	}
}

// WithBreakerOption set the circuit breaker options.
func WithBreakerOption(opts ...circuitbreaker.Option) CircuitBreakerOption {
	return func(o *circuitBreakerOptions) {
		if len(opts) > 0 {
			o.group = group.NewGroup(func() interface{} {
				return circuitbreaker.NewBreaker(opts...)
			})
		}
	}
}

// WithValidCode http code to mark failed
func WithValidCode(code ...int) CircuitBreakerOption {
	return func(o *circuitBreakerOptions) {
		for _, c := range code {
			o.validCodes[c] = struct{}{}
		}
	}
}

// WithDegradeHandler set degrade handler function
func WithDegradeHandler(handler func(c *gin.Context)) CircuitBreakerOption {
	return func(o *circuitBreakerOptions) {
		o.degradeHandler = handler
	}
}

// CircuitBreaker a circuit breaker middleware
func CircuitBreaker(opts ...CircuitBreakerOption) gin.HandlerFunc {
	o := defaultCircuitBreakerOptions()
	o.apply(opts...)

	return func(c *gin.Context) {
		breaker := o.group.Get(c.FullPath()).(circuitbreaker.CircuitBreaker)
		if err := breaker.Allow(); err != nil {
			// NOTE: when client reject request locally, keep adding counter let the drop ratio higher.
			breaker.MarkFailed()
			if o.degradeHandler != nil {
				o.degradeHandler(c)
			} else {
				response.Output(c, http.StatusServiceUnavailable, err.Error())
			}
			c.Abort()
			return
		}

		c.Next()

		code := c.Writer.Status()
		// NOTE: need to check internal and service unavailable error
		_, isHit := o.validCodes[code]
		if isHit {
			breaker.MarkFailed()
		} else {
			breaker.MarkSuccess()
		}
	}
}
