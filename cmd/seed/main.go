package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"ams-ai/internal/config"
	"ams-ai/internal/domain"
	"ams-ai/internal/repository/postgres"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		exit(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	store, err := postgres.New(ctx, cfg.DatabaseURL)
	if err != nil {
		exit(err)
	}
	defer store.Close()

	if err := store.UpsertSeedUser(ctx, "admin@example.com", "admin123", "Demo Admin", domain.RoleAdmin); err != nil {
		exit(err)
	}
	fmt.Println("seed complete")
}

func exit(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
