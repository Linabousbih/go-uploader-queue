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
	dataStore := store.New(db)
	server := apiserver.New(config, logger, dataStore)

	if err := server.Start(ctx); err != nil {
		return err
	}

	return nil
}
