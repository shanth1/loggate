// file: cmd/loggate/main.go
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	goconfig "github.com/shanth1/go-common/config"
	"github.com/shanth1/go-common/logger"
	"github.com/shanth1/loggate/internal/adapters/input/udp"
	"github.com/shanth1/loggate/internal/adapters/output/console"
	"github.com/shanth1/loggate/internal/config"
	"github.com/shanth1/loggate/internal/core/ports"
	"github.com/shanth1/loggate/internal/core/service"
)

func main() {
	// --- Сonfig ---

	cfg := &config.Config{}
	if err := goconfig.Load(goconfig.GetConfigPath(), cfg); err != nil {
		log.Fatalf("load config: %v", err)
	}

	logger.GetLogger("loggate", -1)

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
			log.Fatalf("unknown storage type: %s", storageCfg.Type)
		}

		if storage != nil {
			storages = append(storages, storage)
		}
	}

	if len(storages) == 0 {
		log.Fatal("FATAL: no active storages cofigured")
	}

	// --- Core ---
	logService := service.NewLogService(storages)

	// --- Input/Driver Adapter ---
	udpListener, err := udp.New(cfg.Server.ListenAddress, logService)
	if err != nil {
		log.Fatalf("FATAL: failed to create UDP listener: %v", err)
	}

	// --- Graceful Shutdown ---
	ctx, cancel := context.WithCancel(context.Background())

	go udpListener.Start(ctx)

	// TODO: HTTP server for Prometheus metrics

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("INFO: Shutting down server...")
	cancel()

	for _, s := range storages {
		if err := s.Close(); err != nil {
			log.Printf("ERROR: close storage: %v", err)
		}
	}

	log.Println("INFO: Server gracefully stopped.")
}
