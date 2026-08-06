package main

import (
	"async/config"
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

func main() {
	// Load the Shared AWS Configuration (~/.aws/config)
	ctx := context.Background()
	sdkConfig, err := awsconfig.LoadDefaultConfig(ctx)
	if err != nil {
		log.Fatal(err)
	}

	conf, err := config.New()
	if err != nil {
		log.Fatal(err)
	}

	// Create an Amazon S3 service client
	s3client := s3.NewFromConfig(sdkConfig, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(conf.S3LoacalStackEndpoint)
		o.UsePathStyle = true

	})

	// Get the first page of results for ListObjectsV2 for a bucket
	output, err := s3client.ListBuckets(ctx, &s3.ListBucketsInput{})
	if err != nil {
		log.Fatal(err)
	}

	log.Println("first page results")
	for _, object := range output.Buckets {
		log.Printf(*object.Name)
	}
	sqsClient := sqs.NewFromConfig(sdkConfig, func(options *sqs.Options) {
		options.BaseEndpoint = aws.String(conf.LocalStackEndpoint)
	})

	sqsOutput, err := sqsClient.ListQueues(ctx, &sqs.ListQueuesInput{})
	if err != nil {
		log.Fatal(err)
	}

	for _, q := range sqsOutput.QueueUrls {
		fmt.Println(q)
	}
}
