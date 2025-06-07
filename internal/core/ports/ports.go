// file: internal/core/ports/ports.go
package ports

import (
	"context"

	"github.com/shanth1/loggate/internal/core/domain"
)

type LogIngester interface {
	Ingest(ctx context.Context, msg domain.LogMessage)
}

type LogStorage interface {
	Store(ctx context.Context, messages []domain.LogMessage) error
	Close() error
}
