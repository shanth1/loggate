// file: cmd/loggen/main.go
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/shanth1/loggate/cmd/loggen/internal/config"
	"github.com/shanth1/loggate/cmd/loggen/internal/worker"
)

func main() {
	log.Println("INFO: Starting Log Generator...")

	// 1. Загружаем конфигурацию
	cfg, err := config.Load("cmd/loggen/config/config.yaml")
	if err != nil {
		log.Fatalf("FATAL: failed to load config: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup

	// 2. Запускаем воркеры
	for i := 0; i < cfg.Load.Workers; i++ {
		wg.Add(1)
		go worker.Start(ctx, &wg, i+1, cfg)
	}

	// 3. Ждем сигнала на завершение
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("INFO: Shutting down log generator...")
	cancel()  // Отправляем сигнал на остановку всем воркерам
	wg.Wait() // Ждем, пока все воркеры корректно завершатся

	log.Println("INFO: Log generator stopped.")
}
