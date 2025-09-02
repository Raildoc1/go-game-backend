package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// HTTPServerConfig contains configuration for the optional HTTP server.
type HTTPServerConfig struct {
	// Address is the address the HTTP server listens on.
	Address string `yaml:"address"`
	// ShutdownTimeout defines how long the server has to gracefully
	// shutdown.
	ShutdownTimeout time.Duration `yaml:"shutdown-timeout"`
	// ReadHeaderTimeout limits how long the server waits to read request headers.
	// Helps prevent Slowloris attacks. Usually set to 1–5s.
	ReadHeaderTimeout time.Duration `yaml:"read-header-timeout"`
}

type httpServerSetup struct {
	Cfg            *HTTPServerConfig
	HandlerFactory func() http.Handler
}

// GRPCServerConfig contains configuration for the optional gRPC server.
type GRPCServerConfig struct {
	// Address is the address the gRPC server listens on.
	Address string `yaml:"address"`
}

type grpcServerSetup struct {
	Cfg             *GRPCServerConfig
	SetupServerFunc func(*grpc.Server)
}

// Service orchestrates the lifecycle of application components such as HTTP
// and gRPC servers.
type Service struct {
	httpServerSetup *httpServerSetup
	grpcServerSetup *grpcServerSetup
	goFuncs         []func(context.Context) error
}

func newService(
	httpServerSetup *httpServerSetup,
	grpcServerSetup *grpcServerSetup,
	goFuncs []func(context.Context) error,
) *Service {
	return &Service{
		httpServerSetup: httpServerSetup,
		grpcServerSetup: grpcServerSetup,
		goFuncs:         goFuncs,
	}
}

// Run starts the configured service components and blocks until one of them
// returns an error or the context is cancelled. It attempts a graceful
// shutdown within the given timeout.
func (s *Service) Run(rootCtx context.Context, shutdownTimeout time.Duration) error {
	syscallCtx, cancel := signal.NotifyContext(
		rootCtx,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
		syscall.SIGABRT,
	)
	defer cancel()

	g, errGroupCtx := errgroup.WithContext(syscallCtx)

	done := make(chan struct{})
	context.AfterFunc(errGroupCtx, func() {
		select {
		case <-time.After(shutdownTimeout):
			log.Fatal("failed to gracefully shutdown the server")
		case <-done:
		}
	})

	if s.httpServerSetup != nil {
		httpServer := &http.Server{
			Addr:              s.httpServerSetup.Cfg.Address,
			ReadHeaderTimeout: s.httpServerSetup.Cfg.ReadHeaderTimeout,
			Handler:           s.httpServerSetup.HandlerFactory(),
		}

		g.Go(
			func() error {
				if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
					return fmt.Errorf("http server error: %w", err)
				}
				return nil
			},
		)

		g.Go(func() error {
			<-errGroupCtx.Done()

			// We intentionally decouple shutdown from the (already-canceled) root context.
			// A fresh context with a timeout lets the server drain in-flight requests.
			ctx, cancel := context.WithTimeout(context.Background(), s.httpServerSetup.Cfg.ShutdownTimeout)
			defer cancel()

			//nolint:contextcheck // shutdown must *not* inherit canceled parent
			if err := httpServer.Shutdown(ctx); err != nil {
				return fmt.Errorf("failed to shutdown server: %w", err)
			}
			return nil
		})
	}

	if s.grpcServerSetup != nil {
		grpcServer := grpc.NewServer()
		s.grpcServerSetup.SetupServerFunc(grpcServer)

		g.Go(func() error {
			lis, err := net.Listen("tcp", s.grpcServerSetup.Cfg.Address)
			if err != nil {
				return fmt.Errorf("grpc server failed to start listen: %w", err)
			}
			reflection.Register(grpcServer)
			if err := grpcServer.Serve(lis); err != nil {
				return fmt.Errorf("grpc server error: %w", err)
			}
			return nil
		})

		g.Go(func() error {
			<-errGroupCtx.Done()
			grpcServer.GracefulStop()
			return nil
		})
	}

	for _, f := range s.goFuncs {
		fn := f
		g.Go(func() error { return fn(errGroupCtx) })
	}

	err := g.Wait()
	close(done)
	if err != nil {
		return fmt.Errorf("one of the root goroutines finished with error: %w", err)
	}
	return nil
}
