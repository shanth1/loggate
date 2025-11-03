package app

import (
	"context"
	"fmt"

	"github.com/shanth1/gotools/log"
	"github.com/shanth1/loggate/internal/adapters/input/server"
	"github.com/shanth1/loggate/internal/adapters/input/udp"
	"github.com/shanth1/loggate/internal/adapters/output/console"
	"github.com/shanth1/loggate/internal/config"
	"github.com/shanth1/loggate/internal/core/ports"
	"github.com/shanth1/loggate/internal/core/service"
)

func Run(ctx, shutdownCtx context.Context, cfg *config.Config) {
	logger := log.FromContext(ctx)

	// --- Output/Driven Adapters ---

	// TODO: refactor:
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
			logger.Warn().Msg(fmt.Sprintf("storage type '%s' is enabled but not implemented, skipping", storageCfg.Type))
		}

		if storage != nil {
			storages[storageName] = storage
		}
	}

	if len(storages) == 0 {
		logger.Fatal().Msg("no active storages configured")
	}

	// --- Core ---
	logService := service.NewLogService(storages, cfg.RoutingRules, cfg.DefaultDestinations, cfg.Performance)

	// --- Input/Driver Adapter ---
	infoServer := server.New(cfg.Server.InfoAddress)

	udpListener, err := udp.New(cfg.Server.LogAddress, logService)
	if err != nil {
		logger.Fatal().Err(err).Msg("new udp adapter")
	}

	// --- Start ---
	go logService.Start(ctx)
	go udpListener.Start(ctx)
	go infoServer.Start(ctx)

	// --- Graceful Shutdown ---
	<-ctx.Done()

	logger.Info().Msg("shutting down server...")

	for name, s := range storages {
		if err := s.Close(); err != nil {
			logger.Error().Err(err).Str("storage", name).Msg("close storage")
		}
	}

	logger.Info().Msg("server gracefully stopped")
}
