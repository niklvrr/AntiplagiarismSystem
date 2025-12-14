package main

import (
	"analysis-service/internal/config"
	"analysis-service/internal/infrastructure/minio"
	"analysis-service/internal/infrastructure/pgdb"
	"analysis-service/internal/transport"
	"analysis-service/internal/usecase"
	pb "analysis-service/pkg/api"
	"analysis-service/pkg/logger"
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func main() {
	ctx := context.Background()

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	appLogger, err := logger.NewLogger(cfg.Logger.Level)
	if err != nil {
		log.Fatal(err)
	}
	defer appLogger.Sync()

	db, err := pgdb.NewDatabase(ctx, cfg.Database.URL, appLogger)
	if err != nil {
		appLogger.Fatal("failed to connect to database", zap.Error(err))
	}
	defer db.Close()
	appLogger.Info("database init success")

	minioClient, err := minio.NewClient(ctx, &cfg.Minio)
	if err != nil {
		appLogger.Fatal("failed to connect to minio", zap.Error(err))
	}
	appLogger.Info("minio init success")

	comparator := usecase.NewTextComparator()

	repo := pgdb.NewAnalysisRepository(db)
	service := usecase.NewAnalysisService(repo, minioClient, comparator)
	handler := transport.NewAnalysisHandler(service)

	listener, err := net.Listen("tcp", fmt.Sprintf(":%s", cfg.App.GrpcPort))
	if err != nil {
		appLogger.Fatal("failed to listen", zap.Error(err))
	}

	server := grpc.NewServer()
	pb.RegisterAnalysisServiceServer(server, handler)

	go func() {
		appLogger.Info("starting gRPC server", zap.String("addr", listener.Addr().String()))
		if err := server.Serve(listener); err != nil {
			appLogger.Fatal("failed to serve", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	appLogger.Info("shutting down server")

	server.GracefulStop()

	appLogger.Info("server exited")
}
