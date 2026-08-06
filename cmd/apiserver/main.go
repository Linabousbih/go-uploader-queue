package main

import (
	"async/apiserver"
	"async/config"
	"async/store"
	"context"
	"log"
	"log/slog"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
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
	db, err := store.NewPostgresDB(config)

	if err != nil {
		return err
	}
	sdkConfig, err := awsconfig.LoadDefaultConfig(ctx)
	if err != nil {
		log.Fatal(err)
	}
	dataStore := store.New(db)
	sqsClient := sqs.NewFromConfig(sdkConfig, func(o *sqs.Options) {
		o.BaseEndpoint = aws.String(config.LocalStackEndpoint)
	})

	jwtManager := apiserver.NewJwtManager(config)
	server := apiserver.New(config, logger, dataStore, jwtManager, sqsClient)

	if err := server.Start(ctx); err != nil {
		return err
	}

	return nil
}
