package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"ams-ai/internal/assets"
	"ams-ai/internal/auth"
	"ams-ai/internal/categories"
	"ams-ai/internal/config"
	"ams-ai/internal/documents"
	"ams-ai/internal/httpapi"
	"ams-ai/internal/platform/postgres"
	"ams-ai/internal/platform/storage"
	"ams-ai/internal/reminders"
	"ams-ai/internal/reports"
	"ams-ai/internal/vehicles"
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	cfg, err := config.Load()
	if err != nil {
		log.Error("load config", "error", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	store, err := postgres.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Error("connect database", "error", err)
		os.Exit(1)
	}
	defer store.Close()

	objects, err := storage.NewMinIO(cfg.Storage)
	if err != nil {
		log.Error("create object storage", "error", err)
		os.Exit(1)
	}
	if err := objects.EnsureBucket(ctx); err != nil {
		log.Error("ensure storage bucket", "error", err)
		os.Exit(1)
	}
	if err := store.RegenerateReminders(context.Background(), cfg.ReminderWindowDays); err != nil {
		log.Warn("reminder regeneration skipped", "error", err)
	}

	authSvc := auth.NewService(store, cfg.AuthSecret, cfg.TokenTTL, nil)
	assetSvc := assets.NewService(store, cfg.ReminderWindowDays)
	categorySvc := categories.NewService(store)
	vehicleSvc := vehicles.NewService(store, assetSvc)
	documentSvc := documents.NewService(store, assetSvc, objects, cfg.Storage.MaxUploadBytes, nil)
	reminderSvc := reminders.NewService(store, cfg.ReminderWindowDays)
	reportSvc := reports.NewService(assetSvc, vehicleSvc)
	server := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           httpapi.New(cfg, httpapi.Dependencies{Auth: authSvc, Categories: categorySvc, Assets: assetSvc, Vehicles: vehicleSvc, Documents: documentSvc, Reminders: reminderSvc, Reports: reportSvc}, log),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      90 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	go func() {
		log.Info("api listening", "addr", server.Addr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("api server failed", "error", err)
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Error("api shutdown failed", "error", err)
		os.Exit(1)
	}
	log.Info("api stopped")
}
