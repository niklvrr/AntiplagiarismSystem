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
	"go.uber.org/zap"
)

type AnalysisService interface {
	AnalyseTask(ctx context.Context, taskId uuid.UUID, objectKey string) (bool, error)
	GetReport(ctx context.Context, taskId uuid.UUID) (*domain.Report, error)
	GenerateWordCloud(ctx context.Context, fileContent []byte) (string, error)
}

type AnalysisHandler struct {
	pb.UnimplementedAnalysisServiceServer
	svc    AnalysisService
	logger *zap.Logger
}

func NewAnalysisHandler(svc AnalysisService, logger *zap.Logger) *AnalysisHandler {
	return &AnalysisHandler{
		svc:    svc,
		logger: logger,
	}
}

func (h *AnalysisHandler) AnalyseTask(ctx context.Context, request *pb.AnalyzeTaskRequest) (*pb.AnalyseTaskResponse, error) {
	h.logger.Info("analyse task gRPC request",
		zap.String("task_id", request.TaskId),
		zap.String("object_key", request.ObjectKey))

	taskId, err := uuid.Parse(request.TaskId)
	if err != nil {
		h.logger.Warn("invalid task_id UUID",
			zap.String("task_id", request.TaskId),
			zap.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	status, err := h.svc.AnalyseTask(ctx, taskId, request.ObjectKey)
	if err != nil {
		h.logger.Error("analyse task failed",
			zap.String("task_id", request.TaskId),
			zap.Error(err))
		return nil, mapError(err)
	}

	h.logger.Info("analyse task success",
		zap.String("task_id", request.TaskId),
		zap.Bool("status", status))

	return &pb.AnalyseTaskResponse{
		Status: status,
	}, nil
}

func (h *AnalysisHandler) GetReport(ctx context.Context, request *pb.GetReportRequest) (*pb.GetReportResponse, error) {
	h.logger.Info("get report gRPC request", zap.String("task_id", request.TaskId))

	taskId, err := uuid.Parse(request.TaskId)
	if err != nil {
		h.logger.Warn("invalid task_id UUID",
			zap.String("task_id", request.TaskId),
			zap.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	report, err := h.svc.GetReport(ctx, taskId)
	if err != nil {
		h.logger.Error("get report failed",
			zap.String("task_id", request.TaskId),
			zap.Error(err))
		return nil, mapError(err)
	}

	h.logger.Info("get report success",
		zap.String("task_id", request.TaskId),
		zap.Bool("is_plagiarism", report.IsPlagiarism),
		zap.Float64("plagiarism_percentage", report.PlagiarismPercentage))

	return &pb.GetReportResponse{
		TaskId:               report.TaskId.String(),
		IsPlagiarism:         report.IsPlagiarism,
		PlagiarismPercentage: float32(report.PlagiarismPercentage),
	}, nil
}

func (h *AnalysisHandler) GenerateWordCloud(ctx context.Context, request *pb.GenerateWordCloudRequest) (*pb.GenerateWordCloudResponse, error) {
	h.logger.Info("generate word cloud gRPC request",
		zap.Int("content_size", len(request.FileContent)))

	imageURL, err := h.svc.GenerateWordCloud(ctx, request.FileContent)
	if err != nil {
		h.logger.Error("generate word cloud failed", zap.Error(err))
		return nil, mapError(err)
	}

	h.logger.Info("generate word cloud success",
		zap.String("image_url", imageURL))

	return &pb.GenerateWordCloudResponse{
		ImageUrl: imageURL,
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
