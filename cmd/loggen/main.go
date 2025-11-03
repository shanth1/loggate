package main

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/shanth1/gotools/log"
	"github.com/shanth1/loggate/cmd/loggen/internal/config"
	"github.com/shanth1/loggate/cmd/loggen/internal/worker"
)

func main() {
	logger := log.New()

	logger.Info().Msg("starting log generator...")

	cfg, err := config.Load("cmd/loggen/config/config.yaml")
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to load config")
	}

	ctx, cancel := context.WithCancel(context.Background())
	ctx = log.NewContext(ctx, logger)
	var wg sync.WaitGroup

	for i := 0; i < cfg.Load.Workers; i++ {
		wg.Add(1)
		go worker.Start(ctx, &wg, i+1, cfg)
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info().Msg("Shutting down log generator...")
	cancel()
	wg.Wait()

	logger.Info().Msg("log generator stopped")
}
