package pgdb

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
	"storing-service/internal/domain"
	"storing-service/internal/infrastucture/dto"
)

const (
	createTaskQuery = `
INSERT INTO tasks (id, filename, uploaded_by, created_at) 
VALUES ($1, $2, $3, $4)
RETURNING id`

	getTaskQuery = `
SELECT filename, uploaded_by, created_at
FROM tasks
WHERE id = $1`
)

type StoringRepository struct {
	db     *pgxpool.Pool
	logger *zap.Logger
}

func NewStoringRepository(db *pgxpool.Pool, logger *zap.Logger) *StoringRepository {
	return &StoringRepository{
		db:     db,
		logger: logger,
	}
}

func (r *StoringRepository) CreateTask(ctx context.Context, dto *dto.CreateTaskDTO) (*domain.TaskMetadata, error) {
	r.logger.Debug("executing create task query",
		zap.String("task_id", dto.Id.String()),
		zap.String("filename", dto.FileName))

	err := r.db.QueryRow(ctx, createTaskQuery,
		dto.Id,
		dto.FileName,
		dto.UploadedBy,
		dto.CreatedAt).Scan(&dto.Id)

	if err != nil {
		r.logger.Error("create task query failed",
			zap.String("task_id", dto.Id.String()),
			zap.Error(err))
		return nil, handleDBError(err)
	}

	r.logger.Debug("task created in database", zap.String("task_id", dto.Id.String()))

	return &domain.TaskMetadata{
		Id:         dto.Id,
		Filename:   dto.FileName,
		UploadedBy: dto.UploadedBy,
		CreatedAt:  dto.CreatedAt,
	}, nil
}

func (r *StoringRepository) GetTask(ctx context.Context, dto *dto.GetTaskDTO) (*domain.TaskMetadata, error) {
	r.logger.Debug("executing get task query", zap.String("task_id", dto.Id.String()))

	task := &domain.TaskMetadata{}
	err := r.db.QueryRow(ctx, getTaskQuery, dto.Id).Scan(
		&task.Filename,
		&task.UploadedBy,
		&task.CreatedAt,
	)
	task.Id = dto.Id

	if err != nil {
		r.logger.Error("get task query failed",
			zap.String("task_id", dto.Id.String()),
			zap.Error(err))
		return nil, handleDBError(err)
	}

	r.logger.Debug("task retrieved from database", zap.String("task_id", dto.Id.String()))

	return task, nil
}
