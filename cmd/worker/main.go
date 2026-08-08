package main

import (
	"async/config"
	"async/reports"
	"async/store"
	"context"
	"log"
	"log/slog"
	"os"
	"os/signal"

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
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	conf, err := config.New()
	if err != nil {
		return err
	}
	awsConfig, err := awsconfig.LoadDefaultConfig(ctx)
	if err != nil {
		return err
	}

	db, err := store.NewPostgresDB(conf)
	dataStore := store.New(db)

	s3client := s3.NewFromConfig(awsConfig, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(conf.S3LoacalStackEndpoint)
		o.UsePathStyle = true

	})

	sqsClient := sqs.NewFromConfig(awsConfig, func(options *sqs.Options) {
		options.BaseEndpoint = aws.String(conf.LocalStackEndpoint)
	})

	loz := new(reports.Loz{
		Name: "something",
	})
	builder := reports.NewReportBuilder(dataStore.ReportStore, loz, s3client)
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	worker := reports.NewWorker(conf, logger, builder, sqsClient, 2)
	if err := worker.Start(ctx); err != nil {
		return err
	}
	return nil

}
