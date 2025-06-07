// file: internal/adapters/output/console/storage.go
package console

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/shanth1/loggate/internal/core/domain"
)

type Storage struct{}

func New() *Storage { return &Storage{} }

func (s *Storage) Store(_ context.Context, messages []domain.LogMessage) error {
	for _, msg := range messages {
		raw, _ := json.Marshal(msg)
		fmt.Printf("[CONSOLE STORAGE]: %s\n", string(raw))
	}
	return nil
}

func (s *Storage) Close() error { return nil }
