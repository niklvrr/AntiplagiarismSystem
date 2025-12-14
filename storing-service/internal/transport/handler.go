package transport

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"storing-service/internal/domain"
	"storing-service/internal/errdefs"
	pb "storing-service/pkg/api"
)

type StoringService interface {
	UploadTask(ctx context.Context, filename string, uploadedBy uuid.UUID) (*domain.Task, error)
	GetTask(ctx context.Context, fileId uuid.UUID) (*domain.Task, error)
}

type StoringHandler struct {
	pb.UnimplementedStoringServiceServer
	svc StoringService
}

func NewStoringHandler(service StoringService) *StoringHandler {
	return &StoringHandler{
		svc: service,
	}
}

func (h *StoringHandler) UploadTask(ctx context.Context, request *pb.UploadTaskRequest) (*pb.UploadTaskResponse, error) {
	uploadedBy, err := uuid.Parse(request.UploadedBy)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}
	res, err := h.svc.UploadTask(ctx, request.Filename, uploadedBy)
	if err != nil {
		return nil, mapError(err)
	}

	return &pb.UploadTaskResponse{
		FileId:    res.Id.String(),
		UploadUrl: res.Url,
	}, nil
}

func (h *StoringHandler) GetTask(ctx context.Context, request *pb.GetTaskRequest) (*pb.GetTaskResponse, error) {
	fileId, err := uuid.Parse(request.FileId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}
	res, err := h.svc.GetTask(ctx, fileId)
	if err != nil {
		return nil, mapError(err)
	}

	return &pb.GetTaskResponse{
		FileId:     res.Id.String(),
		Filename:   res.Filename,
		Url:        res.Url,
		UploadedBy: res.UploadedBy.String(),
		UploadedAt: res.CreatedAt.String(),
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
