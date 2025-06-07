package worker

import (
	"context"
	"encoding/json"
	"log"
	"math/rand"
	"net"
	"sync"
	"time"

	"github.com/shanth1/loggate/cmd/loggen/internal/config"
	"github.com/shanth1/loggate/cmd/loggen/internal/generator"
)

func Start(ctx context.Context, wg *sync.WaitGroup, id int, cfg *config.Config) {
	defer wg.Done()
	log.Printf("[Worker %d] Starting...", id)

	var generators []*generator.Generator
	for i := range cfg.Templates {
		generators = append(generators, generator.New(&cfg.Templates[i]))
	}

	if len(generators) == 0 {
		log.Printf("[Worker %d] No templates found, exiting.", id)
		return
	}

	conn, err := net.Dial("udp", cfg.Target)
	if err != nil {
		log.Printf("[Worker %d] ERROR: Could not connect to %s: %v", id, cfg.Target, err)
		return
	}
	defer conn.Close()

	interval := time.Second / time.Duration(cfg.Load.RPS)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Printf("[Worker %d] Stopping.", id)
			return
		case <-ticker.C:
			if cfg.Load.Jitter > 0 {
				time.Sleep(time.Duration(rand.Int63n(int64(cfg.Load.Jitter))))
			}

			gen := generators[rand.Intn(len(generators))]
			msg := gen.Generate()

			payload, _ := json.Marshal(msg)
			_, err := conn.Write(payload)
			if err != nil {
				log.Printf("[Worker %d] WARN: could not send log: %v", id, err)
			}
		}
	}
}
