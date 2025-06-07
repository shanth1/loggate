// file: cmd/loggate/main.go
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/shanth1/loggate/internal/adapters/input/udp"
	"github.com/shanth1/loggate/internal/adapters/output/console"
	"github.com/shanth1/loggate/internal/core/ports"
	"github.com/shanth1/loggate/internal/core/service"
)

func main() {
	// --- Сonfig ---
	// TODO: config

	// --- Output/Driven Adapters ---

	storage := console.New()

	// --- Core ---
	logService := service.NewLogService([]ports.LogStorage{storage})

	// --- Input/Driver Adapter ---
	udpListener, err := udp.New(":10514", logService)
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

	log.Println("INFO: Server gracefully stopped.")
}
