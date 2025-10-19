package middleware

import (
	"bytes"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

var (
	// Print body max length
	defaultMaxLength = 300

	// default zap log
	defaultLogger, _ = zap.NewProduction()

	// Ignore route list
	defaultIgnoreRoutes = map[string]struct{}{
		"/ping":   {},
		"/pong":   {},
		"/health": {},
	}

	// Print error by specified codes
	printErrorBySpecifiedCodes = map[int]bool{
		http.StatusInternalServerError: true,
		http.StatusBadGateway:          true,
		http.StatusServiceUnavailable:  true,
	}

	emptyBody   = []byte("")
	contentMark = []byte(" ...... ")
)

// Option set the gin logger options.
type Option func(*options)

func defaultOptions() *options {
	return &options{
		maxLength:     defaultMaxLength,
		log:           defaultLogger,
		ignoreRoutes:  defaultIgnoreRoutes,
		requestIDFrom: 0,
	}
}

type options struct {
	maxLength     int
	log           *zap.Logger
	ignoreRoutes  map[string]struct{}
	requestIDFrom int // 0: ignore, 1: from context, 2: from header
}

func (o *options) apply(opts ...Option) {
	for _, opt := range opts {
		opt(o)
	}
}

// WithMaxLen logger content max length
func WithMaxLen(maxLen int) Option {
	return func(o *options) {
		if maxLen < len(contentMark) {
			panic("maxLen should be greater than or equal to 8")
		}
		o.maxLength = maxLen
	}
}

// WithLog set log
func WithLog(log *zap.Logger) Option {
	return func(o *options) {
		if log != nil {
			o.log = log
		}
	}
}

// WithIgnoreRoutes no logger content routes
func WithIgnoreRoutes(routes ...string) Option {
	return func(o *options) {
		for _, route := range routes {
			o.ignoreRoutes[route] = struct{}{}
		}
	}
}

// WithPrintErrorByCodes set print error by specified codes
func WithPrintErrorByCodes(code ...int) Option {
	return func(o *options) {
		for _, c := range code {
			printErrorBySpecifiedCodes[c] = true
		}
	}
}

// WithRequestIDFromContext name is field in context, default value is request_id
func WithRequestIDFromContext() Option {
	return func(o *options) {
		o.requestIDFrom = 1
	}
}

// WithRequestIDFromHeader name is field in header, default value is X-Request-Id
func WithRequestIDFromHeader() Option {
	return func(o *options) {
		o.requestIDFrom = 2
	}
}

// ------------------------------------------------------------------------------------------

type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// If there is sensitive information in the body, you can use WithIgnoreRoutes set the route to ignore logging
func getResponseBody(buf *bytes.Buffer, maxLen int) []byte {
	l := buf.Len()
	if l == 0 {
		return []byte("")
	} else if l > maxLen {
		l = maxLen
	}

	body := make([]byte, l)
	n, _ := buf.Read(body)
	if n == 0 {
		return emptyBody
	} else if n < maxLen {
		return body[:n-1]
	}
	return append(body[:maxLen-len(contentMark)], contentMark...)
}

// If there is sensitive information in the body, you can use WithIgnoreRoutes set the route to ignore logging
func getRequestBody(buf *bytes.Buffer, maxLen int) []byte {
	l := buf.Len()
	if l == 0 {
		return []byte("")
	} else if l < maxLen {
		return buf.Bytes()
	}

	body := make([]byte, maxLen)
	copy(body, buf.Bytes())
	return append(body[:maxLen-len(contentMark)], contentMark...)
}

// Logging print request and response info
func Logging(opts ...Option) gin.HandlerFunc {
	o := defaultOptions()
	o.apply(opts...)

	return func(c *gin.Context) {
		start := time.Now()

		// ignore printing of the specified route
		if _, ok := o.ignoreRoutes[c.Request.URL.Path]; ok {
			c.Next()
			return
		}

		buf := bytes.Buffer{}
		_, _ = buf.ReadFrom(c.Request.Body)
		sizeField := zap.Skip()
		bodyField := zap.Skip()
		if c.Request.Method == http.MethodPost || c.Request.Method == http.MethodPut ||
			c.Request.Method == http.MethodPatch || c.Request.Method == http.MethodDelete {
			sizeField = zap.Int("size", buf.Len())
			bodyField = zap.ByteString("body", getRequestBody(&buf, o.maxLength))
		}

		reqID := ""
		if o.requestIDFrom == 1 {
			if v, isExist := c.Get(ContextRequestIDKey); isExist {
				if requestID, ok := v.(string); ok {
					reqID = requestID
				}
			}
		} else if o.requestIDFrom == 2 {
			reqID = c.Request.Header.Get(HeaderXRequestIDKey)
		}
		reqIDField := zap.Skip()
		if reqID != "" {
			reqIDField = zap.String(ContextRequestIDKey, reqID)
		}

		// print input information before processing
		o.log.Info("<<<<",
			zap.String("method", c.Request.Method),
			zap.String("url", c.Request.URL.String()),
			sizeField,
			bodyField,
			reqIDField,
		)

		c.Request.Body = io.NopCloser(&buf)

		// replace writer
		newWriter := &bodyLogWriter{body: &bytes.Buffer{}, ResponseWriter: c.Writer}
		c.Writer = newWriter

		// processing requests
		c.Next()

		// print response message after processing
		httpCode := c.Writer.Status()
		fields := []zap.Field{
			zap.Int("code", httpCode),
			zap.String("method", c.Request.Method),
			zap.String("url", c.Request.URL.Path),
			zap.Int64("time_us", time.Since(start).Microseconds()),
			zap.Int("size", newWriter.body.Len()),
			zap.ByteString("body", getResponseBody(newWriter.body, o.maxLength)),
			reqIDField,
		}
		if printErrorBySpecifiedCodes[httpCode] {
			o.log.WithOptions(zap.AddStacktrace(zap.PanicLevel)).Error(">>>>", fields...)
		} else {
			o.log.Info(">>>>", fields...)
		}
	}
}

// SimpleLog print response info
func SimpleLog(opts ...Option) gin.HandlerFunc {
	o := defaultOptions()
	o.apply(opts...)

	return func(c *gin.Context) {
		start := time.Now()

		// ignore printing of the specified route
		if _, ok := o.ignoreRoutes[c.Request.URL.Path]; ok {
			c.Next()
			return
		}

		reqID := ""
		if o.requestIDFrom == 1 {
			if v, isExist := c.Get(ContextRequestIDKey); isExist {
				if requestID, ok := v.(string); ok {
					reqID = requestID
				}
			}
		} else if o.requestIDFrom == 2 {
			reqID = c.Request.Header.Get(HeaderXRequestIDKey)
		}
		reqIDField := zap.Skip()
		if reqID != "" {
			reqIDField = zap.String(ContextRequestIDKey, reqID)
		}

		// processing requests
		c.Next()

		// print return message after processing
		httpCode := c.Writer.Status()
		fields := []zap.Field{
			zap.Int("code", httpCode),
			zap.String("method", c.Request.Method),
			zap.String("url", c.Request.URL.String()),
			zap.Int64("time_us", time.Since(start).Microseconds()),
			zap.Int("size", c.Writer.Size()),
			reqIDField,
		}
		if printErrorBySpecifiedCodes[httpCode] {
			o.log.WithOptions(zap.AddStacktrace(zap.PanicLevel)).Error("Gin response", fields...)
		} else {
			o.log.Info("Gin response", fields...)
		}
	}
}
