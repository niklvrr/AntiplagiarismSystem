package usecase

import (
	"analysis-service/internal/domain"
	"analysis-service/internal/infrastructure/dto"
	"analysis-service/internal/infrastructure/minio"
	"analysis-service/internal/infrastructure/wordcloud"
	"context"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"time"
)

type AnalysisRepository interface {
	CreateReport(ctx context.Context, dto *dto.CreateReportDTO) error
	GetReport(ctx context.Context, dto *dto.GetReportsDTO) (*domain.Report, error)
}

type FileComparator interface {
	CompareFiles(ctx context.Context, file1, file2 []byte) (float64, error)
}

type AnalysisService struct {
	repo        AnalysisRepository
	minioClient *minio.Client
	comparator  FileComparator
	logger      *zap.Logger
}

func NewAnalysisService(repo AnalysisRepository, client *minio.Client, comparator FileComparator, logger *zap.Logger) *AnalysisService {
	return &AnalysisService{
		repo:        repo,
		minioClient: client,
		comparator:  comparator,
		logger:      logger,
	}
}

func (s *AnalysisService) AnalyseTask(ctx context.Context, taskId uuid.UUID, objectKey string) (bool, error) {
	s.logger.Info("starting task analysis",
		zap.String("task_id", taskId.String()),
		zap.String("object_key", objectKey))

	s.logger.Debug("fetching all file keys from MinIO")
	allKeys, err := s.minioClient.GetAllKeys(ctx)
	if err != nil {
		s.logger.Error("failed to get all keys from MinIO",
			zap.String("task_id", taskId.String()),
			zap.Error(err))
		return false, err
	}

	s.logger.Debug("filtering keys",
		zap.Int("total_keys", len(allKeys)),
		zap.String("target_key", objectKey))
	otherKeys := filterKeys(allKeys, objectKey)
	s.logger.Debug("keys filtered",
		zap.Int("other_keys_count", len(otherKeys)))

	s.logger.Debug("fetching target file from MinIO", zap.String("object_key", objectKey))
	targetFile, err := s.minioClient.GetFile(ctx, objectKey)
	if err != nil {
		s.logger.Error("failed to get target file from MinIO",
			zap.String("object_key", objectKey),
			zap.Error(err))
		return false, err
	}
	s.logger.Debug("target file fetched",
		zap.String("object_key", objectKey),
		zap.Int("file_size", len(targetFile)))

	maxPlagiarism := 0.0
	s.logger.Debug("comparing with other files", zap.Int("files_to_compare", len(otherKeys)))
	for i, key := range otherKeys {
		s.logger.Debug("comparing with file",
			zap.Int("index", i+1),
			zap.Int("total", len(otherKeys)),
			zap.String("other_key", key))

		otherFile, err := s.minioClient.GetFile(ctx, key)
		if err != nil {
			s.logger.Warn("failed to get file for comparison",
				zap.String("key", key),
				zap.Error(err))
			continue
		}

		percentage, err := s.comparator.CompareFiles(ctx, targetFile, otherFile)
		if err != nil {
			s.logger.Warn("failed to compare files",
				zap.String("key", key),
				zap.Error(err))
			continue
		}

		s.logger.Debug("comparison result",
			zap.String("key", key),
			zap.Float64("similarity_percentage", percentage))

		if percentage > maxPlagiarism {
			maxPlagiarism = percentage
		}
	}

	isPlagiarism := false
	const plagiarismThreshold = 50.0
	if maxPlagiarism >= plagiarismThreshold {
		isPlagiarism = true
	}

	s.logger.Info("analysis completed",
		zap.String("task_id", taskId.String()),
		zap.Float64("max_plagiarism", maxPlagiarism),
		zap.Bool("is_plagiarism", isPlagiarism),
		zap.Float64("threshold", plagiarismThreshold))

	dto := &dto.CreateReportDTO{
		TaskId:               taskId,
		IsPlagiarism:         isPlagiarism,
		PlagiarismPercentage: maxPlagiarism,
		CreatedAt:            time.Now(),
	}

	s.logger.Debug("saving report to database", zap.String("task_id", taskId.String()))
	err = s.repo.CreateReport(ctx, dto)
	if err != nil {
		s.logger.Error("failed to create report in database",
			zap.String("task_id", taskId.String()),
			zap.Error(err))
		return false, err
	}

	s.logger.Info("report saved successfully", zap.String("task_id", taskId.String()))
	return true, nil
}

func (s *AnalysisService) GetReport(ctx context.Context, taskId uuid.UUID) (*domain.Report, error) {
	s.logger.Info("getting report", zap.String("task_id", taskId.String()))

	dto := &dto.GetReportsDTO{
		TaskId: taskId,
	}

	s.logger.Debug("fetching report from database", zap.String("task_id", taskId.String()))
	report, err := s.repo.GetReport(ctx, dto)
	if err != nil {
		s.logger.Error("failed to get report from database",
			zap.String("task_id", taskId.String()),
			zap.Error(err))
		return nil, err
	}

	s.logger.Info("report retrieved",
		zap.String("task_id", taskId.String()),
		zap.Bool("is_plagiarism", report.IsPlagiarism),
		zap.Float64("plagiarism_percentage", report.PlagiarismPercentage))

	return report, nil
}

func (s *AnalysisService) GenerateWordCloud(ctx context.Context, fileContent []byte) (string, error) {
	s.logger.Info("generating word cloud",
		zap.Int("file_size", len(fileContent)))

	wordCloudClient := wordcloud.NewClient(s.logger)
	imageURL, err := wordCloudClient.GenerateWordCloud(ctx, fileContent)
	if err != nil {
		s.logger.Error("failed to generate word cloud", zap.Error(err))
		return "", err
	}

	s.logger.Info("word cloud generated successfully",
		zap.String("image_url", imageURL))

	return imageURL, nil
}

func filterKeys(allKeys []string, target string) []string {
	filteredKeys := []string{}
	for _, key := range allKeys {
		if key != target {
			filteredKeys = append(filteredKeys, key)
		}
	}
	return filteredKeys
}
