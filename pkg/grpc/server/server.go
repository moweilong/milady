// Package server is generic grpc server-side.
package server

import (
	"fmt"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/moweilong/milady/pkg/grpc/metrics"
)

// RegisterFn register object
type RegisterFn func(srv *grpc.Server)

// ServiceRegisterFn used to register service address to Consul/ETCD/Nacos/Zookeeper...
type ServiceRegisterFn func() error

// Option set server option
type Option func(*options)

type options struct {
	credentials        credentials.TransportCredentials
	unaryInterceptors  []grpc.UnaryServerInterceptor
	streamInterceptors []grpc.StreamServerInterceptor
	serviceRegisterFn  ServiceRegisterFn

	isShowConnections bool
	connectionOptions []metrics.ConnectionOption
}

func defaultServerOptions() *options {
	return &options{}
}

func (o *options) apply(opts ...Option) {
	for _, opt := range opts {
		opt(o)
	}
}

// WithSecure set secure
func WithSecure(credential credentials.TransportCredentials) Option {
	return func(o *options) {
		o.credentials = credential
	}
}

// WithUnaryInterceptor set unary interceptor
func WithUnaryInterceptor(interceptors ...grpc.UnaryServerInterceptor) Option {
	return func(o *options) {
		o.unaryInterceptors = interceptors
	}
}

// WithStreamInterceptor set stream interceptor
func WithStreamInterceptor(interceptors ...grpc.StreamServerInterceptor) Option {
	return func(o *options) {
		o.streamInterceptors = interceptors
	}
}

// WithServiceRegister set service register
func WithServiceRegister(fn ServiceRegisterFn) Option {
	return func(o *options) {
		o.serviceRegisterFn = fn
	}
}

// WithStatConnections enable stat connections
func WithStatConnections(opts ...metrics.ConnectionOption) Option {
	return func(o *options) {
		o.isShowConnections = true
		o.connectionOptions = opts
	}
}

func customInterceptorOptions(o *options) []grpc.ServerOption {
	var opts []grpc.ServerOption

	if o.credentials != nil {
		opts = append(opts, grpc.Creds(o.credentials))
	}

	if len(o.unaryInterceptors) > 0 {
		option := grpc.ChainUnaryInterceptor(o.unaryInterceptors...)
		opts = append(opts, option)
	}
	if len(o.streamInterceptors) > 0 {
		option := grpc.ChainStreamInterceptor(o.streamInterceptors...)
		opts = append(opts, option)
	}

	return opts
}

// Run grpc server with options, registerFn is the function to register object to the server
func Run(port int, registerFn RegisterFn, options ...Option) (*grpc.Server, error) {
	o := defaultServerOptions()
	o.apply(options...)

	// listening on TCP port
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}

	if o.isShowConnections {
		listener = metrics.NewCustomListener(listener, o.connectionOptions...)
	}

	// create a grpc server where interceptors can be injected
	srv := grpc.NewServer(customInterceptorOptions(o)...)

	// register object to the server
	registerFn(srv)

	// register service address to Consul/ETCD/Nacos/Zookeeper...
	if o.serviceRegisterFn != nil {
		if err = o.serviceRegisterFn(); err != nil {
			return nil, err
		}
	}

	go func() {
		// run the server
		err = srv.Serve(listener)
		if err != nil {
			panic(err)
		}
	}()

	return srv, nil
}
