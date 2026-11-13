package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/hashicorp/go-multierror"

	"github.com/vishenosik/notify/pkg/logs"
)

type App struct {
	di         *diContainer
	httpServer *http.Server

	closed atomic.Bool
}

func New() *App {
	di := newDIContainer()

	slog.SetDefault(di.Logger())
	slog.Debug("loaded config", slog.Any("config", di.Config()))
	_ = di.GRPCServer()

	return &App{
		di: di,
		httpServer: &http.Server{
			Addr:              di.Config().HTTP.Addr.String(),
			Handler:           di.HTTPHandler(),
			ReadHeaderTimeout: 5 * time.Second,
			ReadTimeout:       15 * time.Second,
			IdleTimeout:       60 * time.Second,
		},
	}
}

func (a *App) Run() {
	grpcServer := a.di.GRPCServer()
	lis, err := net.Listen("tcp", a.di.Config().GRPC.Addr.String())
	if err != nil {
		slog.Error("listen grpc", logs.Error(err))
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func(ctx context.Context) {
		a.runBroker(ctx)
	}(ctx)

	go func() {
		if err := a.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("http server failed", logs.Error(err))
			os.Exit(1)
		}
	}()

	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			slog.Error("grpc server failed", logs.Error(err))
			os.Exit(1)
		}
	}()

	a.gracefulShutdown(cancel)
}

//nolint:contextcheck
func (a *App) runBroker(ctx context.Context) {
	a.di.PushSendBroker()

	if err := a.di.RabbitmqClient().Run(ctx); err != nil {
		slog.Error("rabbitmq client failed", logs.Error(err))
		os.Exit(1)
	}
}

func (a *App) gracefulShutdown(cancel context.CancelFunc) {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	slog.Info("shutting down")
	cancel()

	a.di.GRPCServer().GracefulStop()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := a.httpServer.Shutdown(shutdownCtx); err != nil {
		slog.Error("http shutdown failed", logs.Error(err))
	}

	if err := a.Close(shutdownCtx); err != nil {
		slog.Error("close app failed", logs.Error(err))
	}

	slog.Info("app closed")
}

func (a *App) Close(ctx context.Context) error {
	if a.closed.Load() {
		return fmt.Errorf("app already closed")
	}
	defer a.closed.Store(true)

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	errChan := make(chan error)

	go func() {
		defer close(errChan)
		var errs *multierror.Error

		for _, closer := range a.di.closers {
			if err := closer.Close(); err != nil {
				errs = multierror.Append(errs, fmt.Errorf("close component: %w", err))
			}
		}

		errChan <- errs.ErrorOrNil()
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()

	case err := <-errChan:
		return err
	}
}
