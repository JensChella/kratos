package grpc

import (
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/middleware/recovery"

	"google.golang.org/grpc"
)

// ClientOption is gRPC client option.
type ClientOption func(o *clientOptions)

// WithContext with client context.
func WithContext(ctx context.Context) ClientOption {
	return func(c *clientOptions) {
		c.ctx = ctx
	}
}

// WithTimeout with client timeout.
func WithTimeout(timeout time.Duration) ClientOption {
	return func(c *clientOptions) {
		c.timeout = timeout
	}
}

// WithInsecure with client insecure.
func WithInsecure() ClientOption {
	return func(c *clientOptions) {
		c.insecure = true
	}
}

// WithMiddleware with server middleware.
func WithMiddleware(m middleware.Middleware) ClientOption {
	return func(c *clientOptions) {
		c.middleware = m
	}
}

// WithOptions with gRPC options.
func WithOptions(opts ...grpc.DialOption) ClientOption {
	return func(c *clientOptions) {
		c.grpcOpts = opts
	}
}

type clientOptions struct {
	ctx        context.Context
	insecure   bool
	timeout    time.Duration
	middleware middleware.Middleware
	grpcOpts   []grpc.DialOption
}

// NewClient new a grpc transport client.
func NewClient(target string, opts ...ClientOption) (*grpc.ClientConn, error) {
	options := clientOptions{
		ctx:        context.Background(),
		timeout:    500 * time.Millisecond,
		insecure:   false,
		middleware: recovery.Recovery(),
	}
	for _, o := range opts {
		o(&options)
	}
	var grpcOpts = []grpc.DialOption{
		grpc.WithTimeout(options.timeout),
		grpc.WithUnaryInterceptor(UnaryClientInterceptor(options.middleware)),
	}
	if options.insecure {
		grpcOpts = append(grpcOpts, grpc.WithInsecure())
	}
	if len(options.grpcOpts) > 0 {
		grpcOpts = append(grpcOpts, options.grpcOpts...)
	}
	return grpc.DialContext(options.ctx, target, grpcOpts...)
}

// UnaryClientInterceptor retruns a unary client interceptor.
func UnaryClientInterceptor(m middleware.Middleware) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		h := func(ctx context.Context, req interface{}) (interface{}, error) {
			return reply, invoker(ctx, method, req, reply, cc, opts...)
		}
		if m != nil {
			h = m(h)
		}
		_, err := h(ctx, req)
		return err
	}
}
