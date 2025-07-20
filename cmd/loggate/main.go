package main

import (
	"context"
	"fmt"
	"os/signal"
	"syscall"

	"github.com/shanth1/loggate/internal/adapters/input/server"
	"github.com/shanth1/loggate/internal/adapters/input/udp"
	"github.com/shanth1/loggate/internal/adapters/output/console"
	"github.com/shanth1/loggate/internal/common"
	"github.com/shanth1/loggate/internal/config"
	"github.com/shanth1/loggate/internal/core/ports"
	"github.com/shanth1/loggate/internal/core/service"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// --- Сonfig ---
	cfg := config.MustGetConfig()

	logger := common.GetLogger()
	ctx = logger.WithContext(ctx)

	// --- Output/Driven Adapters ---

	storages := make(map[string]ports.LogStorage)

	for storageName, storageCfg := range cfg.Storages {
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
			storages[storageName] = storage
		}
	}

	if len(storages) == 0 {
		logger.Fatal().Msg("no active storages cofigured")
	}

	// --- Core ---
	logService := service.NewLogService(storages, cfg.RoutingRules, cfg.DefaultDestinations)
	logService.Start(ctx)

	// --- Input/Driver Adapter ---
	udpListener, err := udp.New(cfg.Server.LogAddress, logService)
	if err != nil {
		logger.Fatal().Err(err).Msg("new udp adapter")
	}

	server := server.New(cfg.Server.InfoAddress)

	go udpListener.Start(ctx)
	go server.Start(ctx)

	// --- Graceful Shutdown ---

	<-ctx.Done()

	logger.Info().Msg("shutting down server...")

	logService.Shutdown(ctx)

	for _, s := range storages {
		if err := s.Close(); err != nil {
			logger.Error().Err(err).Msg("close storage")
		}
	}

	logger.Info().Msg("server gracefully stopped")
}
