package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"

	"golang.org/x/sync/errgroup"

	"github.com/skryde/booking-check/server/internal/platform/queue"
)

func main() {
	// Run NATS Server
	_queue, err := queue.RunEmbeddedNATS(false, false)
	if err != nil {
		slog.Error("failed to run embedded NATS server", slog.Any("error", err))
		panic(err)
	}

	// Wait for NATS Server shutdown.
	defer _queue.WaitForShutdown()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancel()

	deps, err := buildDependencies(ctx, _queue)
	if err != nil {
		slog.Error("failed to build dependencies", slog.Any("error", err))
		os.Exit(1)
	}

	// Do things before fully app shutdown.
	defer deps.tearDown()

	errGroup, ctx := errgroup.WithContext(ctx)
	onCtxDone := func(f func()) {
		<-ctx.Done()
		f()
	}

	errGroup.Go(func() error { return deps.bot.Start(ctx) })
	errGroup.Go(func() error {
		mux := &http.ServeMux{}
		server := &http.Server{Addr: ":8080", Handler: mux}

		mux.HandleFunc("/subs", deps.api.GetSubscriptions)

		go onCtxDone(func() {
			if err := server.Shutdown(ctx); err != nil {
				slog.Error("failed to shutdown server", slog.Any("error", err))
			}
		})

		return server.ListenAndServe()
	})

	if err := errGroup.Wait(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		slog.Error("error running services", slog.Any("error", err))
	}
}
