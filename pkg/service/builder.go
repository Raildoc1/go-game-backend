package service

import (
	"context"
	"net/http"

	"google.golang.org/grpc"
)

// Config contains common configuration parameters for services built by
// Builder.
type Config struct {
	// Version is the version string of the service.
	Version string `yaml:"version"`
}

// Builder helps in constructing a Service with optional components such as
// initialization functions and HTTP/GRPC servers.
type Builder struct {
	initFunc        func(context.Context) error
	httpServerSetup *httpServerSetup
	grpcServerSetup *grpcServerSetup
}

// NewBuilder creates a new empty Builder instance.
func NewBuilder() *Builder {
	return &Builder{}
}

// WithInitialization registers a function to run during service startup.
func (b *Builder) WithInitialization(initFunc func(context.Context) error) *Builder {
	b.initFunc = initFunc
	return b
}

// WithHTTPServer configures the service to start an HTTP server using the
// provided configuration and handler factory.
func (b *Builder) WithHTTPServer(
	cfg *HTTPServerConfig,
	handlerFactory func() http.Handler,
) *Builder {
	b.httpServerSetup = &httpServerSetup{
		Cfg:            cfg,
		HandlerFactory: handlerFactory,
	}
	return b
}

// WithGRPCServer configures the service to start a gRPC server using the
// provided configuration and setup function.
func (b *Builder) WithGRPCServer(
	cfg *GRPCServerConfig,
	setupServerFunc func(*grpc.Server),
) *Builder {
	b.grpcServerSetup = &grpcServerSetup{
		Cfg:             cfg,
		SetupServerFunc: setupServerFunc,
	}
	return b
}

// Build constructs a Service based on the options configured on the Builder.
func (b *Builder) Build() *Service {
	return newService(
		b.initFunc,
		b.httpServerSetup,
		b.grpcServerSetup,
	)
}
