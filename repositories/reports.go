package repositories

import (
	"context"
	"database/sql"
	"fmt"

	"async/models"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type ReportStore struct {
	db *sqlx.DB
}

func NewReportStore(db *sql.DB) *ReportStore {
	return &ReportStore{
		db: sqlx.NewDb(db, "postgres"),
	}
}

func (s *ReportStore) Create(ctx context.Context, userId uuid.UUID, reportType string) (*models.Report, error) {
	const insert = `INSERT INTO reports (user_id, report_type) VALUED ($1, $2) RETURNING *`
	var report models.Report
	if err := s.db.GetContext(ctx, &report, insert, reportType); err != nil {
		return nil, fmt.Errorf("failed to insert report for user %w", err)
	}

	return &report, nil
}

func (s *ReportStore) Update(ctx context.Context, report *models.Report) (*models.Report, error) {
	const update = `UPDATE reports SET
					output_file_path= $1,
					download_url=$2
					download_url_expires_at=$3
					error_message=$4
					started_at=$5
					completed_at=$6
					failed_at=$7
				WHERE user_id=$8 AND id=$9 RETURNING *`
	var updated models.Report
	if err := s.db.GetContext(ctx, &updated, update,
		report.OutputFilePath,
		report.DownloadUrl,
		report.DownloadUrlExpiresAt,
		report.ErrorMessage,
		report.StartedAt,
		report.CompletedAt,
		report.FailedAt,
		report.UserId,
		report.Id); err != nil {
		return nil, fmt.Errorf("failed to update report %s for user %d: %w", report.Id, report.UserId, err)
	}
	return report, nil
}

func (s *ReportStore) ByPrimaryKey(ctx context.Context, userId, id uuid.UUID) (*models.Report, error) {
	const query = `SELECT 1 FROM reports WHERE user_id=$1 AND id=$2`
	var report models.Report
	if err := s.db.GetContext(ctx, &report, query, userId, id); err != nil {
		return nil, fmt.Errorf("failed to get report %w", err)
	}

	return &report, nil
}
