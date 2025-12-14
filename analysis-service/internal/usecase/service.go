package usecase

import (
	"analysis-service/internal/domain"
	"analysis-service/internal/infrastructure/dto"
	"analysis-service/internal/infrastructure/minio"
	"context"
	"github.com/google/uuid"
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
}

func NewAnalysisService(repo AnalysisRepository, client *minio.Client, comparator FileComparator) *AnalysisService {
	return &AnalysisService{
		repo:        repo,
		minioClient: client,
		comparator:  comparator,
	}
}

func (s *AnalysisService) AnalyseTask(ctx context.Context, taskId uuid.UUID, objectKey string) (bool, error) {
	allKeys, err := s.minioClient.GetAllKeys(ctx)
	if err != nil {
		return false, err
	}

	otherKeys := filterKeys(allKeys, objectKey)

	targetFile, err := s.minioClient.GetFile(ctx, objectKey)
	if err != nil {
		return false, err
	}

	maxPlagiarism := 0.0
	for _, key := range otherKeys {
		otherFile, err := s.minioClient.GetFile(ctx, key)
		if err != nil {
			continue
		}

		percentage, err := s.comparator.CompareFiles(ctx, targetFile, otherFile)
		if err != nil {
			continue
		}

		maxPlagiarism = max(percentage, maxPlagiarism)
	}

	isPlagiarism := false
	const plagiarismThreshold = 50.0
	if maxPlagiarism >= plagiarismThreshold {
		isPlagiarism = true
	}
	dto := &dto.CreateReportDTO{
		TaskId:               taskId,
		IsPlagiarism:         isPlagiarism,
		PlagiarismPercentage: maxPlagiarism,
		CreatedAt:            time.Now(),
	}

	err = s.repo.CreateReport(ctx, dto)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (s *AnalysisService) GetReport(ctx context.Context, taskId uuid.UUID) (*domain.Report, error) {
	dto := &dto.GetReportsDTO{
		TaskId: taskId,
	}

	report, err := s.repo.GetReport(ctx, dto)
	if err != nil {
		return nil, err
	}

	return report, nil
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
