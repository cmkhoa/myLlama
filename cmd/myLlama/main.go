package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"myLlama/internal/api"
	"myLlama/internal/buildinfo"
	"myLlama/internal/config"
	"myLlama/internal/logging"
	"myLlama/internal/runner"
)

func main() {
	if err := run(); err != nil {
		slog.Error("myLlama failed", "error", err)
		os.Exit(1)
	}
}

func run() error {
	version := flag.Bool("version", false, "print build information and exit")
	flag.Parse()
	if *version {
		return json.NewEncoder(os.Stdout).Encode(buildinfo.Current())
	}
	if flag.NArg() != 0 {
		return errors.New("myLlama accepts no positional arguments")
	}
	cfg, err := config.FromEnv()
	if err != nil {
		return err
	}
	var destination io.Writer = os.Stdout
	if cfg.LogPath != "" {
		file, err := logging.Open(cfg.LogPath, cfg.LogMaxBytes, cfg.LogBackups)
		if err != nil {
			return err
		}
		defer file.Close()
		destination = file
	}
	logger := slog.New(slog.NewJSONHandler(destination, nil))
	manager := runner.NewManager(cfg.Runner, logger)
	for _, profile := range cfg.Profiles {
		if err := manager.AddProfile(profile); err != nil {
			return err
		}
	}
	handler := api.NewHandlerWithConfig(manager, logger, cfg.API)
	server := &http.Server{Addr: cfg.ListenAddress, Handler: handler, ReadHeaderTimeout: 5 * time.Second}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	errCh := make(chan error, 1)
	go func() {
		logger.Info("myLlama listening", "address", cfg.ListenAddress, "version", buildinfo.Version)
		errCh <- server.ListenAndServe()
	}()
	var serverErr error
	select {
	case <-ctx.Done():
		logger.Info("shutdown signal received")
	case serverErr = <-errCh:
	}
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("HTTP shutdown failed", "error", err)
		_ = server.Close()
	}
	if serverErr != nil && !errors.Is(serverErr, http.ErrServerClosed) {
		return serverErr
	}
	return nil
}
