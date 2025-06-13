package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/shanth1/loggate/internal/adapters/input/udp"
	"github.com/shanth1/loggate/internal/adapters/output/console"
	"github.com/shanth1/loggate/internal/common"
	"github.com/shanth1/loggate/internal/config"
	"github.com/shanth1/loggate/internal/core/ports"
	"github.com/shanth1/loggate/internal/core/service"
)

func main() {
	// --- Сonfig ---

	cfg := config.MustGetConfig()

	logger := common.GetLogger()

	// --- Output/Driven Adapters ---

	var storages []ports.LogStorage

	for _, storageCfg := range cfg.Storages {
		if !storageCfg.Enabled {
			continue
		}

		var storage ports.LogStorage
		switch storageCfg.Type {
		case "console":
			storage = console.New()
		default:
			logger.Fatal().Msg(fmt.Sprintf("unknown storage type: %s", storageCfg.Type))
		}

		if storage != nil {
			storages = append(storages, storage)
		}
	}

	if len(storages) == 0 {
		logger.Fatal().Msg("no active storages cofigured")
	}

	// --- Core ---
	logService := service.NewLogService(storages)

	// --- Input/Driver Adapter ---
	udpListener, err := udp.New(cfg.Server.ListenAddress, logService)
	if err != nil {
		logger.Fatal().Err(err).Msg("new udp adapter")
	}

	// --- Graceful Shutdown ---

	ctx, cancel := context.WithCancel(context.Background())
	ctx = logger.WithContext(ctx)

	go udpListener.Start(ctx)

	// TODO: HTTP server for Prometheus metrics

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info().Msg("shutting down server...")
	cancel()

	for _, s := range storages {
		if err := s.Close(); err != nil {
			logger.Error().Err(err).Msg("close storage")
		}
	}

	logger.Info().Msg("server gracefully stopped")
}
