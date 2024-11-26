package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/rusher2004/canis-lupus-arctos/server"
	"github.com/rusher2004/canis-lupus-arctos/store"
)

func run(ctx context.Context, rs server.RiskStore, addr string) error {
	handler := slog.NewJSONHandler(os.Stdout, nil)
	logger := slog.New(handler)

	srv := server.NewServer(rs, logger)

	// listen for interrupt signal to neatly shutdown server
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	httpServer := &http.Server{
		Addr:     addr,
		Handler:  srv,
		ErrorLog: slog.NewLogLogger(handler, slog.LevelDebug),
	}

	go func() {
		logger.Info("server listening at " + httpServer.Addr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("error listening and serving", "error", err)
		}
	}()

	// use a waitgroup to block until we receive an interrupt signal
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()

		<-ctx.Done()
		logger.Info("shutting down server")

		downCtx := context.Background()
		downCtx, cancel := context.WithTimeout(downCtx, 5*time.Second)
		defer cancel()

		if err := httpServer.Shutdown(downCtx); err != nil {
			logger.Error("error shutting down server", "error", err)
		}
	}()

	wg.Wait()

	return nil
}

func main() {
	ctx := context.Background()

	rs := store.NewMemoryStore()

	if err := run(ctx, rs, ":8080"); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
