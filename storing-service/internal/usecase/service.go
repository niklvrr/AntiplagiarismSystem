package usecase

import (
	"context"
	"fmt"
	"path"
	"storing-service/internal/domain"
	"storing-service/internal/errdefs"
	"storing-service/internal/infrastucture/dto"
	"time"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
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
	minio          *minio.Client
	bucket         string
	analysisClient AnalysisClient
}

func NewStoringService(repo StoringRepository, minio *minio.Client, bucket string, analysisClient AnalysisClient) *StoringService {
	return &StoringService{
		repo:           repo,
		minio:          minio,
		bucket:         bucket,
		analysisClient: analysisClient,
	}
}

func (s *StoringService) UploadTask(ctx context.Context, filename string, uploadedBy uuid.UUID) (*domain.Task, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return nil, fmt.Errorf("failed to generate uuid: %w", errdefs.ErrInvalidArgument)
	}
	extension := path.Ext(filename)
	if extension == "" {
		return nil, fmt.Errorf("invalid file extension: %w", errdefs.ErrInvalidArgument)
	}
	dto := &dto.CreateTaskDTO{
		Id:         id,
		FileName:   filename,
		UploadedBy: uploadedBy,
		CreatedAt:  time.Now(),
	}

	metaData, err := s.repo.CreateTask(ctx, dto)
	if err != nil {
		return nil, err
	}

	objectKey := fmt.Sprintf("%s%s", id.String(), extension)
	uploadUrl, err := s.minio.PresignedPutObject(ctx, s.bucket, objectKey, time.Hour)
	if err != nil {
		return nil, fmt.Errorf("failed to generate upload url: %w", errdefs.ErrUnavailable)
	}

	status, err := s.analysisClient.AnalyseTask(ctx, id.String(), objectKey)
	if err != nil {
		return nil, err
	}

	if status == false {
		return nil, fmt.Errorf("failed to analyse task status: %w", errdefs.ErrUnavailable)
	}

	return &domain.Task{
		Id:         metaData.Id,
		Filename:   metaData.Filename,
		Url:        uploadUrl.String(),
		UploadedBy: metaData.UploadedBy,
		CreatedAt:  metaData.CreatedAt,
	}, nil
}

func (s *StoringService) GetTask(ctx context.Context, fileId uuid.UUID) (*domain.Task, error) {
	dto := &dto.GetTaskDTO{
		Id: fileId,
	}

	metaData, err := s.repo.GetTask(ctx, dto)
	if err != nil {
		return nil, err
	}

	extension := path.Ext(metaData.Filename)
	objectKey := fmt.Sprintf("%s%s", fileId.String(), extension)
	downloadUrl, err := s.minio.PresignedGetObject(ctx, s.bucket, objectKey, time.Hour, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to generate download url: %w", errdefs.ErrUnavailable)
	}

	return &domain.Task{
		Id:         metaData.Id,
		Filename:   metaData.Filename,
		Url:        downloadUrl.String(),
		UploadedBy: metaData.UploadedBy,
		CreatedAt:  metaData.CreatedAt,
	}, nil
}
