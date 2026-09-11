package controllers

import (
	"log/slog"

	"async/config"
	"async/helpers"
	"async/repositories"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

type Controller struct {
	config        *config.Config
	logger        *slog.Logger
	store         *repositories.Store
	jwtManager    *helpers.JwtManager
	sqsClient     *sqs.Client
	presignClient *s3.PresignClient
}

func New(config *config.Config, logger *slog.Logger, store *repositories.Store, jwtManager *helpers.JwtManager, sqsClient *sqs.Client, presignClient *s3.PresignClient) *Controller {

	return &Controller{
		config:        config,
		logger:        logger,
		store:         store,
		jwtManager:    jwtManager,
		sqsClient:     sqsClient,
		presignClient: presignClient,
	}
}
