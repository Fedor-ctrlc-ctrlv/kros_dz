package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"dz_kros/go-service/internal/handlers"
	"dz_kros/go-service/internal/logging"
	"dz_kros/go-service/internal/middleware"
	"dz_kros/go-service/internal/store"
)

func main() {
	logger := logging.New()
	api := &handlers.API{Log: logger, Orders: store.NewOrders()}
	server := &http.Server{Addr: ":8080", Handler: middleware.Logging(logger, api.Routes()), ReadHeaderTimeout: 5 * time.Second}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	go func() {
		logger.Info("go api started", "address", server.Addr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("server failed", "error", err.Error())
			os.Exit(1)
		}
	}()
	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("shutdown failed", "error", err.Error())
	}
	logger.Log(context.Background(), slog.LevelInfo, "server stopped")
}
