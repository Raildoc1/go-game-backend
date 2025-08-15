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

type HTTPServerConfig struct {
	Address         string        `yaml:"address"`
	ShutdownTimeout time.Duration `yaml:"shutdown-timeout"`
}

type httpServerSetup struct {
	Cfg            *HTTPServerConfig
	HandlerFactory func() http.Handler
}

type GRPCServerConfig struct {
	Address string `yaml:"address"`
}

type grpcServerSetup struct {
	Cfg             *GRPCServerConfig
	SetupServerFunc func(*grpc.Server)
}

type DeinitSetupConfig struct {
	ShutdownTimeout time.Duration `xml:"shutdown-timeout"`
}

type deinitSetup struct {
	Cfg        *DeinitSetupConfig
	DeinitFunc func() error
}

type Service struct {
	initFunc        func(context.Context) error
	httpServerSetup *httpServerSetup
	grpcServerSetup *grpcServerSetup
}

func newService(
	initFunc func(context.Context) error,
	httpServerSetup *httpServerSetup,
	grpcServerSetup *grpcServerSetup,
) *Service {
	return &Service{
		initFunc:        initFunc,
		httpServerSetup: httpServerSetup,
		grpcServerSetup: grpcServerSetup,
	}
}

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

	if s.initFunc != nil {
		err := s.initFunc(syscallCtx)
		if err != nil {
			return fmt.Errorf("init service failed: %w", err)
		}
	}

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
			Addr:    s.httpServerSetup.Cfg.Address,
			Handler: s.httpServerSetup.HandlerFactory(),
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

	err := g.Wait()
	close(done)
	return err
}
