package dto

import (
	"github.com/google/uuid"
	"time"
)

type CreateTaskDTO struct {
	Id         uuid.UUID
	FileName   string
	UploadedBy uuid.UUID
	CreatedAt  time.Time
}

type GetTaskDTO struct {
	Id uuid.UUID
}
