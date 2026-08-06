package reports

import (
	"async/config"
	"async/store"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/csv"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
)

type ReportBuilder struct {
	config      *config.Config
	reportStore *store.ReportStore
	lozClient   *Loz
	s3Client    *s3.Client
}

func NewReportBuilder(reportStore *store.ReportStore, lozClient *Loz, s3Client *s3.Client) *ReportBuilder {
	return &ReportBuilder{
		reportStore: reportStore,
		lozClient:   lozClient,
		s3Client:    s3Client,
	}
}

func (b *ReportBuilder) Build(ctx context.Context, userId, reportId uuid.UUID) (report *store.Report, err error) {
	report, err = b.reportStore.ByPrimaryKey(ctx, userId, reportId)
	if err != nil {
		return nil, fmt.Errorf("failed to get report %w", err)
	}

	if report.StartedAt != nil {
		return report, nil
	}

	defer func() {
		if err != nil {
			failedTime := time.Now()
			report.FailedAt = &failedTime
			report.ErrorMessage = aws.String(err.Error())
			if _, updateErr := b.reportStore.Update(ctx, report); updateErr != nil {
				fmt.Errorf("failed to update report %w", err)
			}

		}
	}()

	now := time.Now()
	report.StartedAt = &now
	report.CompletedAt, report.FailedAt, report.ErrorMessage, report.DownloadUrl, report.DownloadUrlExpiresAt, report.OutputFilePath = nil, nil, nil, nil, nil, nil

	report, err = b.reportStore.Update(ctx, report)
	if err != nil {
		return nil, fmt.Errorf("failed to update report %w", err)
	}
	resp, err := b.lozClient.GenerateReport(ctx, report)
	if err != nil {
		return nil, fmt.Errorf("failed to generate report %w", err)
	}
	resp, err = b.lozClient.GetMonsters()
	if err != nil {
		return nil, fmt.Errorf("failed to get monsters %w", err)
	}
	if len(resp.Data) == 0 {
		return nil, fmt.Errorf("no monsters found")
	}
	var buffer bytes.Buffer
	gzipWriter := gzip.NewWriter(&buffer)
	csvWriter := csv.NewWriter(gzipWriter)

	header := []string{"name", "id", "category", "description", "image", "common_locations", "drops", "dlc"}
	if err := csvWriter.Write(header); err != nil {
		return nil, fmt.Errorf("failed to write csv header: %w", err)
	}

	// ****** add data ******
	// for _, monster := range resp.Data {
	// 	csvRow:=[]string{
	// 		monster.Name,
	// 		fmt.Sprintf("%d",monster.id)
	// 	}
	// }

	csvWriter.Flush()
	if err := csvWriter.Error(); err != nil {
		return nil, fmt.Errorf("failed to flush writer: %w", err)
	}

	if err := gzipWriter.Close(); err != nil {
		return nil, fmt.Errorf("failed to close gzip writer: %w", err)
	}

	// Create an S3 path
	key := "/users" + userId.String() + "/report" + reportId.String() + "csv.gz"
	_, err = b.s3Client.PutObject(ctx, &s3.PutObjectInput{
		Key:    aws.String(key),
		Bucket: aws.String(b.config.S3Bucket),
		Body:   bytes.NewReader(buffer.Bytes()),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to upload report: %w", err)
	}

	updateTime := time.Now()
	report.OutputFilePath = &key
	report.CompletedAt = &updateTime
	report, err = b.reportStore.Update(ctx, report)
	if err != nil {
		return nil, fmt.Errorf("failed to update report %w", err)
	}
	return report, nil
}
