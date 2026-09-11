package models

import (
	"time"

	"github.com/google/uuid"
)

type Report struct {
	UserId               uuid.UUID  `db:"user_id"`
	Id                   uuid.UUID  `db:"id"`
	ReportType           string     `db:"report_type"`
	OutputFilePath       *string    `db:"output_file_path"`
	DownloadUrl          *string    `db:"download_url"`
	DownloadUrlExpiresAt *time.Time `db:"download_url_expires_at"`
	ErrorMessage         *string    `db:"error_message"`
	CreatedAt            time.Time  `db:"created_at"`
	StartedAt            *time.Time `db:"started_at"`
	CompletedAt          *time.Time `db:"completed_at"`
	FailedAt             *time.Time `db:"failed_at"`
}

func (r *Report) IsDone() bool {
	return r.CompletedAt != nil || r.FailedAt != nil
}

func (r *Report) Status() string {
	switch {
	case r.StartedAt == nil:
		return "requested"
	case r.StartedAt != nil && r.IsDone():
		return "processing"
	case r.CompletedAt != nil:
		return "completed"
	case r.FailedAt != nil:
		return "failed"
	}

	return "unknown"
}
