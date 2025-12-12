package main

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"log"
	"net"
	"storing-service/internal/config"
	"storing-service/internal/infrastucture/minio"
	"storing-service/internal/infrastucture/pgdb"
	"storing-service/internal/transport"
	"storing-service/internal/usecase"
	pb "storing-service/pkg/api"
	"storing-service/pkg/logger"
)

func main() {
	ctx := context.Background()

	// TODO config init
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	// TODO logger init
	logger, err := logger.NewLogger(cfg.Logger.Level)
	if err != nil {
		log.Fatal(err)
	}
	defer logger.Sync()

	// TODO database init
	db, err := pgdb.NewDatabase(ctx, cfg.Database.URL, logger)
	if err != nil {
		logger.Fatal("database init failed", zap.Error(err))
	}
	defer db.Close()
	logger.Info("database init successfully")

	// TODO minio init
	fileStorage, err := minio.NewClient(ctx, &cfg.Minio)
	if err != nil {
		logger.Fatal("minio init failed", zap.Error(err))
	}
	logger.Info("minio file storage init successfully")

	// TODO all layers init
	pgrepo := pgdb.NewStoringRepository(db)
	service := usecase.NewStoringService(pgrepo, fileStorage, cfg.Minio.Bucket)
	handler := transport.NewStoringHandler(service)

	// TODO grpc server init
	listener, err := net.Listen("tcp", fmt.Sprintf(":%s", cfg.App.GrpcPort))
	if err != nil {
		logger.Fatal("grpc listen failed", zap.Error(err))
	}

	grpcServer := grpc.NewServer()
	pb.RegisterStoringServiceServer(grpcServer, handler)
	if err := grpcServer.Serve(listener); err != nil {
		logger.Fatal("grpc serve failed", zap.Error(err))
	}
}
