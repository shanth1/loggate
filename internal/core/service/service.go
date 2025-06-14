package service

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/shanth1/gotools/log"
	"github.com/shanth1/loggate/internal/config"
	"github.com/shanth1/loggate/internal/core/domain"
	"github.com/shanth1/loggate/internal/core/ports"
)

const (
	bufferSize   = 1000
	batchSize    = 1000
	batchTimeout = 5 * time.Second
)

type LogService struct {
	storages            map[string]ports.LogStorage
	destinationChannels map[string]chan domain.LogMessage
	wg                  sync.WaitGroup
	routingRules        []*config.RoutingRule
	defaultDestinations []string
}

func NewLogService(
	storages map[string]ports.LogStorage,
	routingRules []*config.RoutingRule,
	defaultDestinations []string,
) *LogService {
	return &LogService{
		storages:            storages,
		destinationChannels: make(map[string]chan domain.LogMessage),
		routingRules:        routingRules,
		defaultDestinations: defaultDestinations,
	}
}

func (s *LogService) Start(ctx context.Context) {
	logger := log.FromCtx(ctx)
	logger.Info().Msg("starting log service workers")

	for name, storage := range s.storages {
		s.destinationChannels[name] = make(chan domain.LogMessage, bufferSize)
		s.wg.Add(1)
		go s.runWorker(ctx, name, storage, s.destinationChannels[name])
	}
	logger.Info().Msg("log service workers started")
}

func (s *LogService) Shutdown(ctx context.Context) {
	logger := log.FromCtx(ctx)
	logger.Info().Msg("waiting for log service workers to stop...")

	s.wg.Wait()

	logger.Info().Msg("log service workers stopped gracefully")
}

func (s *LogService) Ingest(ctx context.Context, msg domain.LogMessage) {
	logger := log.FromCtx(ctx)

	destinations := s.findDestinations(msg)
	if len(destinations) == 0 {
		destinations = s.defaultDestinations
	}

	for _, destName := range destinations {
		if channel, ok := s.destinationChannels[destName]; ok {
			select {
			case channel <- msg:
			default:
				logger.Warn().Msg(fmt.Sprintf("buffer for destination '%s' is full. Log message dropped.", destName))
			}
		} else {
			logger.Warn().Msg(fmt.Sprintf("routing rule points to a non-configured storage: %s", destName))
		}
	}
}

func (s *LogService) runWorker(ctx context.Context, name string, storage ports.LogStorage, channel <-chan domain.LogMessage) {
	defer s.wg.Done()
	logger := log.FromCtx(ctx).With().Str("storage", name).Logger()

	batch := make([]domain.LogMessage, 0, batchSize)
	ticker := time.NewTicker(batchTimeout)
	defer ticker.Stop()

	logger.Info().Msg("worker started")

	for {
		select {
		case <-ctx.Done():
			logger.Info().Msg("shutdown signal received, flushing remaining logs")

			flushCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()
			if len(batch) > 0 {
				if err := storage.Store(flushCtx, batch); err != nil {
					logger.Error().Err(err).Msg("failed to store final batch")
				}
			}
			logger.Info().Msg("worker stopped")
			return

		case msg, ok := <-channel:
			if !ok {
				logger.Warn().Msg("channel closed unexpectedly, flushing remaining logs")
				if len(batch) > 0 {
					if err := storage.Store(ctx, batch); err != nil {
						logger.Error().Err(err).Msg("failed to store final batch on channel close")
					}
				}
				logger.Info().Msg("worker stopped")
				return
			}

			batch = append(batch, msg)
			if len(batch) >= batchSize {
				if err := storage.Store(ctx, batch); err != nil {
					logger.Error().Err(err).Msg("failed to store batch")
				}
				batch = make([]domain.LogMessage, 0, batchSize)
				ticker.Reset(batchTimeout)
			}

		case <-ticker.C:
			if len(batch) > 0 {
				if err := storage.Store(ctx, batch); err != nil {
					logger.Error().Err(err).Msg("failed to store batch on timeout")
				}
				batch = make([]domain.LogMessage, 0, batchSize)
			}
		}
	}
}

func (s *LogService) findDestinations(msg domain.LogMessage) []string {
	for _, rule := range s.routingRules {
		if rule.MatchCondition.Service != msg.Service {
			continue
		}
		if !strings.EqualFold(rule.MatchCondition.Level, msg.Level) {
			continue
		}
		return rule.Destinations
	}

	return nil
}
