package main

import (
	"api-gateway/internal/config"
	"api-gateway/internal/infrastructure/analysis"
	"api-gateway/internal/infrastructure/storing"
	"api-gateway/internal/transport"
	"api-gateway/pkg/logger"
	"context"
	"fmt"
	"go.uber.org/zap"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	ctx := context.Background()
	cfg := config.LoadConfig()

	appLogger, err := logger.NewLogger(cfg.Logger.Level)
	if err != nil {
		panic(err)
	}
	defer appLogger.Sync()

	analysisClient, err := analysis.NewClient(ctx, cfg.AnalysisService.Endpoint)
	if err != nil {
		appLogger.Fatal("failed to connect to analysis service", zap.Error(err))
	}
	defer analysisClient.Close()

	storingClient, err := storing.NewClient(ctx, cfg.StoringService.Endpoint)
	if err != nil {
		appLogger.Fatal("failed to connect to storing service", zap.Error(err))
	}
	defer storingClient.Close()

	handler := transport.NewHandler(analysisClient, storingClient)
	router := transport.NewRouter(handler, appLogger)

	server := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.App.HTTPPort),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		appLogger.Info("starting HTTP server", zap.String("addr", server.Addr))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			appLogger.Fatal("server failed", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	appLogger.Info("shutting down server")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		appLogger.Error("server forced to shutdown", zap.Error(err))
	}

	appLogger.Info("server exited")
}
