package domain

import (
	"github.com/google/uuid"
	"time"
)

type Task struct {
	Id         uuid.UUID `db:"id"`
	Filename   string    `db:"filename"`
	Url        string    `db:"url"`
	UploadedBy uuid.UUID `db:"uploaded_by"`
	CreatedAt  time.Time `db:"created_at"`
}

type TaskMetadata struct {
	Id         uuid.UUID `db:"id"`
	Filename   string    `db:"filename"`
	UploadedBy uuid.UUID `db:"uploaded_by"`
	CreatedAt  time.Time `db:"created_at"`
}
