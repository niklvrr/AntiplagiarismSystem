package pgdb

import (
	"analysis-service/internal/domain"
	"analysis-service/internal/infrastructure/dto"
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

const (
	createReportQuery = `
INSERT INTO reports (task_id, is_plagiarism, plagiarism_percentage, created_at)
VALUES ($1, $2, $3, $4)
RETURNING task_id`

	getReportQuery = `
SELECT task_id, is_plagiarism, plagiarism_percentage, created_at
FROM reports
WHERE task_id = $1`
)

type AnalysisRepository struct {
	db     *pgxpool.Pool
	logger *zap.Logger
}

func NewAnalysisRepository(db *pgxpool.Pool, logger *zap.Logger) *AnalysisRepository {
	return &AnalysisRepository{
		db:     db,
		logger: logger,
	}
}

func (r *AnalysisRepository) CreateReport(ctx context.Context, dto *dto.CreateReportDTO) error {
	r.logger.Debug("executing create report query",
		zap.String("task_id", dto.TaskId.String()),
		zap.Bool("is_plagiarism", dto.IsPlagiarism),
		zap.Float64("plagiarism_percentage", dto.PlagiarismPercentage))

	err := r.db.QueryRow(ctx, createReportQuery,
		dto.TaskId,
		dto.IsPlagiarism,
		dto.PlagiarismPercentage,
		dto.CreatedAt).Scan(&dto.TaskId)

	if err != nil {
		r.logger.Error("create report query failed",
			zap.String("task_id", dto.TaskId.String()),
			zap.Error(err))
		return handleDBError(err)
	}

	r.logger.Debug("report created in database", zap.String("task_id", dto.TaskId.String()))
	return nil
}

func (r *AnalysisRepository) GetReport(ctx context.Context, dto *dto.GetReportsDTO) (*domain.Report, error) {
	r.logger.Debug("executing get report query", zap.String("task_id", dto.TaskId.String()))

	report := &domain.Report{}
	err := r.db.QueryRow(ctx, getReportQuery, dto.TaskId).Scan(
		&report.TaskId,
		&report.IsPlagiarism,
		&report.PlagiarismPercentage,
		&report.CreatedAt)

	if err != nil {
		r.logger.Error("get report query failed",
			zap.String("task_id", dto.TaskId.String()),
			zap.Error(err))
		return nil, handleDBError(err)
	}

	r.logger.Debug("report retrieved from database", zap.String("task_id", dto.TaskId.String()))
	return report, nil
}
