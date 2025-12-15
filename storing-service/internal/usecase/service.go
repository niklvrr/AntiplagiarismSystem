package usecase

import (
	"context"
	"fmt"
	"io"
	"path"
	"storing-service/internal/domain"
	"storing-service/internal/errdefs"
	"storing-service/internal/infrastucture/dto"
	minio1 "storing-service/internal/infrastucture/minio"
	"time"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"go.uber.org/zap"
)

type StoringRepository interface {
	CreateTask(ctx context.Context, dto *dto.CreateTaskDTO) (*domain.TaskMetadata, error)
	GetTask(ctx context.Context, dto *dto.GetTaskDTO) (*domain.TaskMetadata, error)
}

type AnalysisClient interface {
	AnalyseTask(ctx context.Context, taskId, objectKey string) (bool, error)
}

type StoringService struct {
	repo           StoringRepository
	minio          *minio1.Client
	bucket         string
	analysisClient AnalysisClient
	logger         *zap.Logger
}

func NewStoringService(repo StoringRepository, minio *minio1.Client, bucket string, analysisClient AnalysisClient, logger *zap.Logger) *StoringService {
	return &StoringService{
		repo:           repo,
		minio:          minio,
		bucket:         bucket,
		analysisClient: analysisClient,
		logger:         logger,
	}
}

func (s *StoringService) UploadTask(ctx context.Context, filename string, uploadedBy uuid.UUID) (*domain.Task, error) {
	s.logger.Info("starting upload task",
		zap.String("filename", filename),
		zap.String("uploaded_by", uploadedBy.String()))

	id, err := uuid.NewV7()
	if err != nil {
		s.logger.Error("failed to generate UUID", zap.Error(err))
		return nil, fmt.Errorf("failed to generate uuid: %w", errdefs.ErrInvalidArgument)
	}

	extension := path.Ext(filename)
	if extension == "" {
		s.logger.Warn("invalid file extension", zap.String("filename", filename))
		return nil, fmt.Errorf("invalid file extension: %w", errdefs.ErrInvalidArgument)
	}

	dto := &dto.CreateTaskDTO{
		Id:         id,
		FileName:   filename,
		UploadedBy: uploadedBy,
		CreatedAt:  time.Now(),
	}

	s.logger.Debug("creating task in database", zap.String("task_id", id.String()))
	metaData, err := s.repo.CreateTask(ctx, dto)
	if err != nil {
		s.logger.Error("failed to create task in database",
			zap.String("task_id", id.String()),
			zap.Error(err))
		return nil, err
	}

	s.logger.Debug("task created in database", zap.String("task_id", id.String()))

	objectKey := fmt.Sprintf("%s%s", id.String(), extension)
	s.logger.Debug("generating presigned upload URL",
		zap.String("object_key", objectKey),
		zap.String("bucket", s.bucket))

	uploadUrl, err := s.minio.PresignedPutObject(ctx, objectKey, time.Hour)
	if err != nil {
		s.logger.Error("failed to generate upload URL",
			zap.String("object_key", objectKey),
			zap.Error(err))
		return nil, fmt.Errorf("failed to generate upload url: %w", errdefs.ErrUnavailable)
	}

	s.logger.Info("starting async analysis",
		zap.String("task_id", id.String()),
		zap.String("object_key", objectKey))
	go s.startAnalysisAsync(context.Background(), id.String(), objectKey, metaData.Filename)

	s.logger.Info("upload task completed",
		zap.String("task_id", id.String()),
		zap.String("filename", filename))

	return &domain.Task{
		Id:         metaData.Id,
		Filename:   metaData.Filename,
		Url:        uploadUrl.String(),
		UploadedBy: metaData.UploadedBy,
		CreatedAt:  metaData.CreatedAt,
	}, nil
}

func (s *StoringService) GetTask(ctx context.Context, fileId uuid.UUID) (*domain.Task, error) {
	s.logger.Info("getting task", zap.String("file_id", fileId.String()))

	dto := &dto.GetTaskDTO{
		Id: fileId,
	}

	s.logger.Debug("fetching task from database", zap.String("file_id", fileId.String()))
	metaData, err := s.repo.GetTask(ctx, dto)
	if err != nil {
		s.logger.Error("failed to get task from database",
			zap.String("file_id", fileId.String()),
			zap.Error(err))
		return nil, err
	}

	extension := path.Ext(metaData.Filename)
	objectKey := fmt.Sprintf("%s%s", fileId.String(), extension)

	s.logger.Debug("generating presigned download URL",
		zap.String("object_key", objectKey),
		zap.String("bucket", s.bucket))

	downloadUrl, err := s.minio.PresignedGetObject(ctx, objectKey, time.Hour, nil)
	if err != nil {
		s.logger.Error("failed to generate download URL",
			zap.String("object_key", objectKey),
			zap.Error(err))
		return nil, fmt.Errorf("failed to generate download url: %w", errdefs.ErrUnavailable)
	}

	s.logger.Info("get task completed", zap.String("file_id", fileId.String()))

	return &domain.Task{
		Id:         metaData.Id,
		Filename:   metaData.Filename,
		Url:        downloadUrl.String(),
		UploadedBy: metaData.UploadedBy,
		CreatedAt:  metaData.CreatedAt,
	}, nil
}

func (s *StoringService) GetFileContent(ctx context.Context, fileId uuid.UUID) ([]byte, error) {
	s.logger.Info("getting file content", zap.String("file_id", fileId.String()))

	dto := &dto.GetTaskDTO{
		Id: fileId,
	}

	s.logger.Debug("fetching task metadata from database", zap.String("file_id", fileId.String()))
	metaData, err := s.repo.GetTask(ctx, dto)
	if err != nil {
		s.logger.Error("failed to get task from database",
			zap.String("file_id", fileId.String()),
			zap.Error(err))
		return nil, err
	}

	extension := path.Ext(metaData.Filename)
	objectKey := fmt.Sprintf("%s%s", fileId.String(), extension)

	s.logger.Debug("fetching file from MinIO",
		zap.String("object_key", objectKey),
		zap.String("bucket", s.bucket))

	obj, err := s.minio.GetObject(ctx, objectKey, minio.GetObjectOptions{})
	if err != nil {
		s.logger.Error("failed to get file from MinIO",
			zap.String("object_key", objectKey),
			zap.Error(err))
		return nil, fmt.Errorf("failed to get file from storage: %w", errdefs.ErrUnavailable)
	}
	defer obj.Close()

	content, err := io.ReadAll(obj)
	if err != nil {
		s.logger.Error("failed to read file content",
			zap.String("object_key", objectKey),
			zap.Error(err))
		return nil, fmt.Errorf("failed to read file content: %w", errdefs.ErrUnavailable)
	}

	s.logger.Info("file content retrieved",
		zap.String("file_id", fileId.String()),
		zap.Int("content_size", len(content)))

	return content, nil
}

func (s *StoringService) startAnalysisAsync(ctx context.Context, taskId, objectKey, filename string) {
	maxRetries := 30
	retryInterval := 2 * time.Second
	timeout := 5 * time.Minute

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	for i := 0; i < maxRetries; i++ {
		select {
		case <-ctx.Done():
			s.logger.Warn("analysis timeout waiting for file upload",
				zap.String("task_id", taskId),
				zap.String("object_key", objectKey))
			return
		default:
		}

		exists, err := s.checkFileExists(ctx, objectKey)
		if err != nil {
			s.logger.Warn("failed to check file existence",
				zap.String("task_id", taskId),
				zap.String("object_key", objectKey),
				zap.Error(err))
			time.Sleep(retryInterval)
			continue
		}

		if exists {
			s.logger.Info("file uploaded, starting analysis",
				zap.String("task_id", taskId),
				zap.String("object_key", objectKey))

			status, err := s.analysisClient.AnalyseTask(ctx, taskId, objectKey)
			if err != nil {
				s.logger.Error("failed to start analysis",
					zap.String("task_id", taskId),
					zap.String("object_key", objectKey),
					zap.Error(err))
				return
			}

			if status {
				s.logger.Info("analysis started successfully",
					zap.String("task_id", taskId))
			} else {
				s.logger.Warn("analysis returned false status",
					zap.String("task_id", taskId))
			}
			return
		}

		time.Sleep(retryInterval)
	}

	s.logger.Warn("file not found after max retries, skipping analysis",
		zap.String("task_id", taskId),
		zap.String("object_key", objectKey),
		zap.Int("max_retries", maxRetries))
}

func (s *StoringService) checkFileExists(ctx context.Context, objectKey string) (bool, error) {
	_, err := s.minio.StatObject(ctx, objectKey, minio.StatObjectOptions{})
	if err != nil {
		errResp := minio.ToErrorResponse(err)
		if errResp.Code == "NoSuchKey" || errResp.Code == "NotFound" {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
