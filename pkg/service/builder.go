package service

import (
	"context"
	"google.golang.org/grpc"
	"net/http"
)

type Config struct {
	Version string `yaml:"version"`
}

type Builder struct {
	initFunc        func(context.Context) error
	httpServerSetup *httpServerSetup
	grpcServerSetup *grpcServerSetup
}

func NewBuilder() *Builder {
	return &Builder{}
}

func (b *Builder) WithInitialization(initFunc func(context.Context) error) *Builder {
	b.initFunc = initFunc
	return b
}

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

func (b *Builder) Build() *Service {
	return newService(
		b.initFunc,
		b.httpServerSetup,
		b.grpcServerSetup,
	)
}
