package main

import (
	"async/apiserver"
	"async/config"
	"context"
	"log"
	"log/slog"
	"os"
	"time"
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
	server := apiserver.New(config, logger)

	if err := server.Start(ctx); err != nil {
		return err
	}

	return nil
}
