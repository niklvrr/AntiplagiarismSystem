package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"storing-service/internal/config"
	"storing-service/internal/infrastucture/analysis"
	"storing-service/internal/infrastucture/minio"
	"storing-service/internal/infrastucture/pgdb"
	"storing-service/internal/transport"
	"storing-service/internal/usecase"
	pb "storing-service/pkg/api"
	"storing-service/pkg/logger"
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
		appLogger.Fatal("database init failed", zap.Error(err))
	}
	defer db.Close()
	appLogger.Info("database init successfully")

	fileStorage, err := minio.NewClient(ctx, &cfg.Minio)
	if err != nil {
		appLogger.Fatal("minio init failed", zap.Error(err))
	}
	appLogger.Info("minio file storage init successfully")

	analysisClient, err := analysis.NewClient(ctx, cfg.Analysis.URL)
	if err != nil {
		appLogger.Fatal("analysis init failed", zap.Error(err))
	}
	defer analysisClient.Close()

	pgrepo := pgdb.NewStoringRepository(db, appLogger)
	service := usecase.NewStoringService(pgrepo, fileStorage, cfg.Minio.Bucket, analysisClient, appLogger)
	handler := transport.NewStoringHandler(service, appLogger)

	listener, err := net.Listen("tcp", fmt.Sprintf(":%s", cfg.App.GrpcPort))
	if err != nil {
		appLogger.Fatal("grpc listen failed", zap.Error(err))
	}

	grpcServer := grpc.NewServer()
	pb.RegisterStoringServiceServer(grpcServer, handler)

	go func() {
		appLogger.Info("starting gRPC server", zap.String("addr", listener.Addr().String()))
		if err := grpcServer.Serve(listener); err != nil {
			appLogger.Fatal("grpc serve failed", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	appLogger.Info("shutting down server")

	grpcServer.GracefulStop()

	appLogger.Info("server exited")
}
