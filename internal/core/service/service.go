// file: internal/core/service/log_service.go
package service

import (
	"context"
	"log"

	"github.com/shanth1/loggate/internal/core/domain"
	"github.com/shanth1/loggate/internal/core/ports"
)

type LogService struct {
	storages []ports.LogStorage
}

func NewLogService(storages []ports.LogStorage) *LogService {
	return &LogService{storages: storages}
}

func (s *LogService) Ingest(ctx context.Context, msg domain.LogMessage) {
	batch := []domain.LogMessage{msg}

	for _, storage := range s.storages {
		// Запуск в горутине, чтобы не блокировать прием UDP пакетов
		go func(st ports.LogStorage) {
			if err := st.Store(ctx, batch); err != nil {
				log.Printf("ERROR: failed to store log to %T: %v\n", st, err)
			}
		}(storage)
	}
}
