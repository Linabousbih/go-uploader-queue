package controllers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"async/helpers"
	"async/models"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/google/uuid"
)

func (s *Controller) CreateReport() http.HandlerFunc {
	return helpers.Handler(func(w http.ResponseWriter, r *http.Request) error {
		req, err := helpers.Decode[*models.CreateReportRequest](r)
		if err != nil {
			return helpers.NewErrWithStatus(http.StatusInternalServerError, err)
		}

		user, ok := helpers.UserFromContext(r.Context())
		if !ok {
			return helpers.NewErrWithStatus(http.StatusUnauthorized, err)
		}
		report, err := s.store.ReportStore.Create(r.Context(), user.Id, req.ReportType)

		if err != nil {
			return helpers.NewErrWithStatus(http.StatusInternalServerError, err)
		}

		sqsMessqge := models.SqsMessage{
			UserId:   user.Id,
			ReportId: report.Id,
		}

		bytes, err := json.Marshal(sqsMessqge)
		if err != nil {
			return helpers.NewErrWithStatus(http.StatusInternalServerError, err)
		}

		queueUrlOutput, err := s.sqsClient.GetQueueUrl(r.Context(), &sqs.GetQueueUrlInput{
			QueueName: aws.String(s.config.SqsQueue),
		})

		if err != nil {
			return helpers.NewErrWithStatus(http.StatusInternalServerError, err)
		}

		_, err = s.sqsClient.SendMessage(r.Context(), &sqs.SendMessageInput{
			MessageBody: aws.String(string(bytes)),
			QueueUrl:    queueUrlOutput.QueueUrl,
		})
		if err != nil {
			return helpers.NewErrWithStatus(http.StatusInternalServerError, err)
		}

		if err := helpers.Encode(models.ApiResponse[models.CreateReportResponse]{
			Data: &models.CreateReportResponse{
				Id:                   report.Id,
				UserId:               report.UserId,
				ReportType:           report.ReportType,
				OutputFilePath:       report.OutputFilePath,
				DownloadUrl:          report.DownloadUrl,
				DownloadUrlExpiresAt: report.DownloadUrlExpiresAt,
				ErrorMessage:         report.ErrorMessage,
				Status:               report.Status(),
			},
		}, http.StatusCreated, w); err != nil {
			return helpers.NewErrWithStatus(http.StatusInternalServerError, err)
		}

		return nil

	})
}

func (s *Controller) GetReport() http.HandlerFunc {
	return helpers.Handler(func(w http.ResponseWriter, r *http.Request) error {
		reportIdStr := r.PathValue("id")
		reportId, err := uuid.Parse(reportIdStr)
		if err != nil {
			return helpers.NewErrWithStatus(http.StatusBadRequest, err)
		}
		user, ok := helpers.UserFromContext(r.Context())
		if !ok {
			return helpers.NewErrWithStatus(http.StatusUnauthorized, fmt.Errorf("user not found in context"))
		}

		report, err := s.store.ReportStore.ByPrimaryKey(r.Context(), user.Id, reportId)

		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return helpers.NewErrWithStatus(http.StatusNotFound, err)
			}
			return helpers.NewErrWithStatus(http.StatusInternalServerError, err)
		}

		if report.CompletedAt != nil {
			needsRefresh := report.DownloadUrlExpiresAt == nil || report.DownloadUrlExpiresAt.Before(time.Now())
			if needsRefresh || report.DownloadUrl == nil {
				// Generate a new presigned URL using the S3 Presign Client
				// To S3 Client (presigned client)
				expiresAt := time.Now().Add(10 * time.Second)
				signedUrl, err := s.presignClient.PresignGetObject(r.Context(), &s3.GetObjectInput{
					Bucket: aws.String(s.config.S3Bucket),
					Key:    report.OutputFilePath,
				}, func(o *s3.PresignOptions) {
					o.Expires = time.Second * 10
				})

				if err != nil {
					return helpers.NewErrWithStatus(http.StatusInternalServerError, err)
				}

				report.DownloadUrl = &signedUrl.URL
				report.DownloadUrlExpiresAt = &expiresAt
				report, err = s.store.ReportStore.Update(r.Context(), report)

				if err != nil {
					return helpers.NewErrWithStatus(http.StatusInternalServerError, err)
				}
			}
		}

		if err := helpers.Encode(models.ApiResponse[models.CreateReportResponse]{
			Data: &models.CreateReportResponse{
				Id:                   report.Id,
				UserId:               report.UserId,
				ReportType:           report.ReportType,
				OutputFilePath:       report.OutputFilePath,
				DownloadUrl:          report.DownloadUrl,
				DownloadUrlExpiresAt: report.DownloadUrlExpiresAt,
				ErrorMessage:         report.ErrorMessage,
				Status:               report.Status(),
			},
		}, http.StatusOK, w); err != nil {
			return helpers.NewErrWithStatus(http.StatusInternalServerError, err)
		}
		return nil
	})
}
