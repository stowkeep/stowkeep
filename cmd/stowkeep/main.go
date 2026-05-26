package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/stowkeep/stowkeep/pkg/config"
	"github.com/stowkeep/stowkeep/pkg/db"
	applog "github.com/stowkeep/stowkeep/pkg/observability/log"
	"github.com/stowkeep/stowkeep/pkg/server"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", slog.String("error", err.Error()))
		os.Exit(1)
	}

	logger := applog.New(cfg.LogLevel, cfg.LogFormat, cfg.Version, cfg.LogAddSource)
	slog.SetDefault(logger)

	if cfg.ResolvedDriver() == "sqlite" {
		if err := os.MkdirAll(filepath.Dir(cfg.ResolvedSQLitePath()), 0o750); err != nil {
			logger.Error("failed to create data directory", slog.String("error", err.Error()))
			os.Exit(1)
		}
	}

	database, err := db.Open(cfg)
	if err != nil {
		logger.Error("failed to connect database", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer func() {
		if err := database.Close(); err != nil {
			logger.Error("failed to close database", slog.String("error", err.Error()))
		}
	}()

	migrationsRoot := cfg.MigrationsDir
	if migrationsRoot == "" {
		migrationsRoot = "migrations"
	}
	if err := db.Up(database, cfg.ResolvedDriver(), migrationsRoot); err != nil {
		logger.Error("failed to run migrations", slog.String("error", err.Error()))
		os.Exit(1)
	}

	srv := server.New(cfg, logger, database)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("HTTP server stopped", slog.String("error", err.Error()))
			stop()
		}
	}()

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("shutdown error", slog.String("error", err.Error()))
	}
}
