package pgdb

import (
	"analysis-service/internal/domain"
	"analysis-service/internal/infrastructure/dto"
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	createReportQuery = `
INSERT INTO analysis_reports (task_id, is_plagiarism, plagiarism_percentage, created_at)
VALUES ($1, $2, $3, $4)
RETURNING task_id`

	getReportQuery = `
SELECT task_id, is_plagiarism, plagiarism_percentage, created_at
FROM analysis_reports
WHERE task_id = $1`
)

type AnalysisRepository struct {
	db *pgxpool.Pool
}

func NewAnalysisRepository(db *pgxpool.Pool) *AnalysisRepository {
	return &AnalysisRepository{
		db: db,
	}
}

func (r *AnalysisRepository) CreateReport(ctx context.Context, dto *dto.CreateReportDTO) error {
	err := r.db.QueryRow(ctx, createReportQuery,
		dto.TaskId,
		dto.IsPlagiarism,
		dto.PlagiarismPercentage,
		dto.CreatedAt).Scan(&dto.TaskId)

	if err != nil {
		return handleDBError(err)
	}
	return nil
}

func (r *AnalysisRepository) GetReport(ctx context.Context, dto *dto.GetReportsDTO) (*domain.Report, error) {
	report := &domain.Report{}
	err := r.db.QueryRow(ctx, getReportQuery, dto.TaskId).Scan(
		&report.TaskId,
		&report.IsPlagiarism,
		&report.PlagiarismPercentage,
		&report.CreatedAt)

	if err != nil {
		return nil, handleDBError(err)
	}

	return report, nil
}
