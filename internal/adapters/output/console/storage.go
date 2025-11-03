// file: internal/adapters/output/console/storage.go
package console

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/shanth1/loggate/internal/core/domain"
)

type Storage struct{}

func New() *Storage { return &Storage{} }

func (s *Storage) Store(_ context.Context, messages []domain.LogMessage) error {
	for _, msg := range messages {
		flatMap := make(map[string]interface{}, 5+len(msg.Fields))
		for k, v := range msg.Fields {
			flatMap[k] = v
		}

		flatMap["time"] = msg.Time.Format(time.RFC3339Nano)
		flatMap["level"] = msg.Level
		flatMap["app"] = msg.App
		flatMap["service"] = msg.Service
		flatMap["message"] = msg.Message

		raw, _ := json.Marshal(flatMap)
		fmt.Println(string(raw))
	}
	return nil
}

func (s *Storage) Close() error { return nil }
