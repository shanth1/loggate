package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/shanth1/gotools/log"
	"github.com/shanth1/loggate/internal/config"
	"github.com/shanth1/loggate/internal/core/domain"
	"github.com/shanth1/loggate/internal/core/ports"
)

type LogService struct {
	storages map[string]ports.LogStorage
	rules    []config.RoutingRule
	defaults []string
}

func NewLogService(
	storages map[string]ports.LogStorage,
	rules []config.RoutingRule,
	defaults []string,
) *LogService {
	return &LogService{
		storages: storages,
		rules:    rules,
		defaults: defaults,
	}
}

func (s *LogService) Ingest(ctx context.Context, msg domain.LogMessage) {
	logger := log.FromCtx(ctx)

	destinations := s.findDestinations(msg)
	if len(destinations) == 0 {
		destinations = s.defaults
	}

	batch := []domain.LogMessage{msg}

	for _, destName := range destinations {
		if storage, ok := s.storages[destName]; ok {
			go func(st ports.LogStorage, destName string) {
				if err := st.Store(ctx, batch); err != nil {
					logger.Error().Err(err).Msg(fmt.Sprintf("failed to store log to %s", destName))
				}
			}(storage, destName)
		} else {
			logger.Warn().Msg(fmt.Sprintf("routing rule not found in configured storages: %s", destName))
		}
	}
}

func (s *LogService) findDestinations(msg domain.LogMessage) []string {
	for _, rule := range s.rules {
		isMatch := true
		if rule.MatchCondition.Service != msg.Service {
			isMatch = false
		}
		if !strings.EqualFold(rule.MatchCondition.Level, msg.Level) {
			isMatch = false
		}
		if isMatch {
			return rule.Destinations
		}
	}

	return nil
}
