package transport

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"go.uber.org/zap"
	"storing-service/internal/domain"
	"storing-service/internal/errdefs"
	pb "storing-service/pkg/api"
)

type StoringService interface {
	UploadTask(ctx context.Context, filename string, uploadedBy uuid.UUID) (*domain.Task, error)
	GetTask(ctx context.Context, fileId uuid.UUID) (*domain.Task, error)
	GetFileContent(ctx context.Context, fileId uuid.UUID) ([]byte, error)
}

type StoringHandler struct {
	pb.UnimplementedStoringServiceServer
	svc    StoringService
	logger *zap.Logger
}

func NewStoringHandler(service StoringService, logger *zap.Logger) *StoringHandler {
	return &StoringHandler{
		svc:    service,
		logger: logger,
	}
}

func (h *StoringHandler) UploadTask(ctx context.Context, request *pb.UploadTaskRequest) (*pb.UploadTaskResponse, error) {
	h.logger.Info("upload task gRPC request",
		zap.String("filename", request.Filename),
		zap.String("uploaded_by", request.UploadedBy))

	uploadedBy, err := uuid.Parse(request.UploadedBy)
	if err != nil {
		h.logger.Warn("invalid uploaded_by UUID",
			zap.String("uploaded_by", request.UploadedBy),
			zap.Error(err))
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	res, err := h.svc.UploadTask(ctx, request.Filename, uploadedBy)
	if err != nil {
		h.logger.Error("upload task failed",
			zap.String("filename", request.Filename),
			zap.Error(err))
		return nil, mapError(err)
	}

	h.logger.Info("upload task success",
		zap.String("file_id", res.Id.String()),
		zap.String("filename", request.Filename))

	return &pb.UploadTaskResponse{
		FileId:    res.Id.String(),
		UploadUrl: res.Url,
	}, nil
}

func (h *StoringHandler) GetTask(ctx context.Context, request *pb.GetTaskRequest) (*pb.GetTaskResponse, error) {
	h.logger.Info("get task gRPC request", zap.String("file_id", request.FileId))

	fileId, err := uuid.Parse(request.FileId)
	if err != nil {
		h.logger.Warn("invalid file_id UUID",
			zap.String("file_id", request.FileId),
			zap.Error(err))
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	res, err := h.svc.GetTask(ctx, fileId)
	if err != nil {
		h.logger.Error("get task failed",
			zap.String("file_id", request.FileId),
			zap.Error(err))
		return nil, mapError(err)
	}

	h.logger.Info("get task success", zap.String("file_id", request.FileId))

	return &pb.GetTaskResponse{
		FileId:     res.Id.String(),
		Filename:   res.Filename,
		Url:        res.Url,
		UploadedBy: res.UploadedBy.String(),
		UploadedAt: res.CreatedAt.String(),
	}, nil
}

func (h *StoringHandler) GetFileContent(ctx context.Context, request *pb.GetFileContentRequest) (*pb.GetFileContentResponse, error) {
	h.logger.Info("get file content gRPC request", zap.String("file_id", request.FileId))

	fileId, err := uuid.Parse(request.FileId)
	if err != nil {
		h.logger.Warn("invalid file_id UUID",
			zap.String("file_id", request.FileId),
			zap.Error(err))
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	content, err := h.svc.GetFileContent(ctx, fileId)
	if err != nil {
		h.logger.Error("get file content failed",
			zap.String("file_id", request.FileId),
			zap.Error(err))
		return nil, mapError(err)
	}

	h.logger.Info("get file content success",
		zap.String("file_id", request.FileId),
		zap.Int("content_size", len(content)))

	return &pb.GetFileContentResponse{
		Content: content,
	}, nil
}

func mapError(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, errdefs.ErrNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, errdefs.ErrInvalidArgument):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, errdefs.ErrAlreadyExists):
		return status.Error(codes.AlreadyExists, err.Error())
	case errors.Is(err, errdefs.ErrUnavailable):
		return status.Error(codes.Unavailable, err.Error())
	default:
		return status.Error(codes.Internal, err.Error())
	}
}
