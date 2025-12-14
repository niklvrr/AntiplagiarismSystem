package transport

import (
	"analysis-service/internal/domain"
	"analysis-service/internal/errdefs"
	pb "analysis-service/pkg/api"
	"context"
	"errors"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AnalysisService interface {
	AnalyseTask(ctx context.Context, taskId uuid.UUID, objectKey string) (bool, error)
	GetReport(ctx context.Context, taskId uuid.UUID) (*domain.Report, error)
}

type AnalysisHandler struct {
	pb.UnimplementedAnalysisServiceServer
	svc AnalysisService
}

func NewAnalysisHandler(svc AnalysisService) *AnalysisHandler {
	return &AnalysisHandler{
		svc: svc,
	}
}

func (h *AnalysisHandler) AnalyseTask(ctx context.Context, request *pb.AnalyzeTaskRequest) (*pb.AnalyseTaskResponse, error) {
	taskId, err := uuid.Parse(request.TaskId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	status, err := h.svc.AnalyseTask(ctx, taskId, request.ObjectKey)
	if err != nil {
		return nil, mapError(err)
	}

	return &pb.AnalyseTaskResponse{
		Status: status,
	}, nil
}

func (h *AnalysisHandler) GetReport(ctx context.Context, request *pb.GetReportRequest) (*pb.GetReportResponse, error) {
	taskId, err := uuid.Parse(request.TaskId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	report, err := h.svc.GetReport(ctx, taskId)
	if err != nil {
		return nil, mapError(err)
	}

	return &pb.GetReportResponse{
		TaskId:               report.TaskId.String(),
		IsPlagiarism:         report.IsPlagiarism,
		PlagiarismPercentage: float32(report.PlagiarismPercentage),
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
