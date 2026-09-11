package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"time"

	"async/apiserver"
	"async/config"
	"async/controllers"
	"async/database"
	"async/helpers"
	"async/repositories"
	"async/routes"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}

}

func run() error {
	config, err := config.New()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err != nil {
		return err
	}
	jsonHandler := slog.NewJSONHandler(os.Stdout, nil)
	logger := slog.New(jsonHandler)
	db, err := database.NewPostgresDB(config)

	if err != nil {
		return err
	}
	sdkConfig, err := awsconfig.LoadDefaultConfig(ctx)
	if err != nil {
		log.Fatal(err)
	}
	dataStore := repositories.New(db)
	sqsClient := sqs.NewFromConfig(sdkConfig, func(o *sqs.Options) {
		o.BaseEndpoint = aws.String(config.LocalStackEndpoint)
	})

	s3client := s3.NewFromConfig(sdkConfig, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(config.S3LoacalStackEndpoint)
		o.UsePathStyle = true

	})

	presignClient := s3.NewPresignClient(s3client)

	jwtManager := helpers.NewJwtManager(config)
	controller := controllers.New(config, logger, dataStore, jwtManager, sqsClient, presignClient)
	handler := routes.New(controller, logger, jwtManager, dataStore.Users)
	server := apiserver.New(config, logger, handler)

	if err := server.Start(ctx); err != nil {
		return err
	}

	return nil
}
