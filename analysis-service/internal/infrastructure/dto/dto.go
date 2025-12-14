package dto

import (
	"github.com/google/uuid"
	"time"
)

type CreateReportDTO struct {
	TaskId               uuid.UUID
	IsPlagiarism         bool
	PlagiarismPercentage float64
	CreatedAt            time.Time
}

type GetReportsDTO struct {
	TaskId uuid.UUID
}
