package pgdb

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
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
	db *pgxpool.Pool
}

func NewStoringRepository(db *pgxpool.Pool) *StoringRepository {
	return &StoringRepository{
		db: db,
	}
}

func (r *StoringRepository) CreateTask(ctx context.Context, dto *dto.CreateTaskDTO) (*domain.TaskMetadata, error) {
	err := r.db.QueryRow(ctx, createTaskQuery,
		dto.Id,
		dto.FileName,
		dto.UploadedBy,
		dto.CreatedAt).Scan(&dto.Id)

	if err != nil {
		return nil, handleDBError(err)
	}

	return &domain.TaskMetadata{
		Id:         dto.Id,
		Filename:   dto.FileName,
		UploadedBy: dto.UploadedBy,
		CreatedAt:  dto.CreatedAt,
	}, nil
}

func (r *StoringRepository) GetTask(ctx context.Context, dto *dto.GetTaskDTO) (*domain.TaskMetadata, error) {
	task := &domain.TaskMetadata{}
	err := r.db.QueryRow(ctx, getTaskQuery, dto.Id).Scan(
		&task.Id,
		&task.Filename,
		&task.UploadedBy,
	)

	if err != nil {
		return nil, handleDBError(err)
	}

	return task, nil
}
