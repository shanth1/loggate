package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net"
	"sync"
	"time"

	"github.com/shanth1/gotools/log"
	"github.com/shanth1/loggate/cmd/loggen/internal/config"
	"github.com/shanth1/loggate/cmd/loggen/internal/generator"
)

func Start(ctx context.Context, wg *sync.WaitGroup, id int, cfg *config.Config) {
	logger := log.FromCtx(ctx).With().Int("id", id).Logger()

	defer wg.Done()
	logger.Info().Msg(fmt.Sprintf("[Worker %d] Starting...", id))

	var generators []*generator.Generator
	for i := range cfg.Templates {
		generators = append(generators, generator.New(&cfg.Templates[i]))
	}

	if len(generators) == 0 {
		logger.Warn().Msg("no templates found, exiting")
		return
	}

	conn, err := net.Dial("udp", cfg.Target)
	if err != nil {
		logger.Error().Err(err).Msg(fmt.Sprintf("could not connect to %s", cfg.Target))
		return
	}
	defer conn.Close()

	interval := time.Second / time.Duration(cfg.Load.RPS)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Info().Msg("stopped")
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
				logger.Warn().Err(err).Msg("could not send log")
			}
		}
	}
}
