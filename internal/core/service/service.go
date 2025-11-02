package service

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/shanth1/gotools/log"
	"github.com/shanth1/loggate/internal/config"
	"github.com/shanth1/loggate/internal/core/domain"
	"github.com/shanth1/loggate/internal/core/ports"
)

type LogService struct {
	storages            map[string]ports.LogStorage
	destinationChannels map[string]chan domain.LogMessage
	wg                  sync.WaitGroup
	routingRules        []*config.RoutingRule
	defaultDestinations []string
	performance         *config.Performance
}

func NewLogService(
	storages map[string]ports.LogStorage,
	routingRules []*config.RoutingRule,
	defaultDestinations []string,
	performance *config.Performance,
) *LogService {
	if performance.BufferSize == 0 {
		performance.BufferSize = 1000
	}
	if performance.BatchSize == 0 {
		performance.BatchSize = 1000
	}
	if performance.BatchTimeoutMs == 0 {
		performance.BatchTimeoutMs = 5000
	}

	return &LogService{
		storages:            storages,
		destinationChannels: make(map[string]chan domain.LogMessage),
		routingRules:        routingRules,
		defaultDestinations: defaultDestinations,
		performance:         performance,
	}
}

func (s *LogService) Start(ctx context.Context) {
	logger := log.FromCtx(ctx)
	logger.Info().Msg("starting log service workers")

	for name, storage := range s.storages {
		s.destinationChannels[name] = make(chan domain.LogMessage, s.performance.BufferSize)
		s.wg.Add(1)
		go s.runWorker(ctx, name, storage, s.destinationChannels[name])
	}
	logger.Info().Msg("log service workers started")

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
				logger.Warn().Str("destination", destName).Msg("buffer is full. Log message dropped.")
			}
		} else {
			logger.Warn().Str("destination", destName).Msg("routing rule points to a non-configured or disabled storage")
		}
	}
}

func (s *LogService) runWorker(ctx context.Context, name string, storage ports.LogStorage, channel <-chan domain.LogMessage) {
	defer s.wg.Done()
	logger := log.FromCtx(ctx).With().Str("storage", name).Logger()

	batchSize := s.performance.BatchSize
	batchTimeout := time.Duration(s.performance.BatchTimeoutMs) * time.Millisecond

	batch := make([]domain.LogMessage, 0, batchSize)
	ticker := time.NewTicker(batchTimeout)
	defer ticker.Stop()

	logger.Info().Msg("worker started")

	for {
		select {
		case <-ctx.Done():
			logger.Info().Msg("shutdown signal received, flushing remaining logs")
			for len(channel) > 0 {
				batch = append(batch, <-channel)
			}
			if len(batch) > 0 {
				flushCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				if err := storage.Store(flushCtx, batch); err != nil {
					logger.Error().Err(err).Msg("failed to store final batch")
				}
				cancel()
			}
			logger.Info().Msg("worker stopped")
			return

		case msg := <-channel:
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
		if rule.MatchCondition == nil {
			continue
		}

		if rule.MatchCondition.Service != "" && rule.MatchCondition.Service != msg.Service {
			continue
		}

		if rule.MatchCondition.Level != "" && !strings.EqualFold(rule.MatchCondition.Level, msg.Level) {
			continue
		}

		return rule.Destinations
	}

	return nil
}
