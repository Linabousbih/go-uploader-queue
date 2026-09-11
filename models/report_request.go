package models

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type CreateReportRequest struct {
	ReportType string `json:"report_type"`
}

type CreateReportResponse struct {
	UserId               uuid.UUID  `json:"user_id"`
	Id                   uuid.UUID  `json:"id"`
	ReportType           string     `json:"report_type"`
	OutputFilePath       *string    `json:"output_file_path"`
	DownloadUrl          *string    `json:"download_url"`
	DownloadUrlExpiresAt *time.Time `json:"download_url_expires_at"`
	ErrorMessage         *string    `json:"error_message"`
	CreatedAt            time.Time  `json:"created_at"`
	StartedAt            *time.Time `json:"started_at"`
	CompletedAt          *time.Time `json:"completed_at"`
	FailedAt             *time.Time `json:"failed_at"`
	Status               string     `json:"status"`
}

func (r *CreateReportRequest) Validate() error {
	if r.ReportType == "" {
		return errors.New("report type is required")
	}
	return nil
}
