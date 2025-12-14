package domain

import (
	"github.com/google/uuid"
	"time"
)

type Report struct {
	TaskId               uuid.UUID
	IsPlagiarism         bool
	PlagiarismPercentage float64
	CreatedAt            time.Time
}
