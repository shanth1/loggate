# 5. Development and Extension

[RU](./05-development.ru.md)

This section is for developers who want to contribute to LogGate or extend its functionality.

## 5.1. Project Structure

```
├── cmd/ # Application entry points
│ ├── loggate/ # Main LogGate service
│ └── loggen/ # Log generation utility
├── config/ # Default configuration for LogGate
├── docs/ # Project documentation
├── internal/ # Application source code
│ ├── adapters/ # Adapters (Ports and Adapters pattern)
│ │ ├── input/ # Input adapters (e.g., UDP, TCP)
│ │ └── output/ # Output adapters (e.g., console, clickhouse)
│ ├── app/ # Application start logic
│ ├── common/ # Shared code
│ ├── config/ # Configuration loading and parsing logic
│ └── core/ # Core business logic
│   ├── domain/ # Main data structures (e.g., LogMessage)
│   ├── ports/ # Interfaces for interacting with the core
│   └── service/ # Implementation of the main business logic
├── scripts/ # Administration scripts (backup, restore)
├── Makefile # Utilities for building, running, and managing
├── go.mod / go.sum # Go dependencies
└── docker-compose.yaml # Service stack definition
```

## 5.2. Adding a New Storage (Output Adapter)

LogGate's architecture makes it easy to add support for new log storage systems. To do this, you need to implement an adapter that satisfies the `ports.LogStorage` interface.

### Step 1: Implement the `LogStorage` Interface

The interface is defined in `internal/core/ports/ports.go`:

```go
type LogStorage interface {
    // Store saves a batch of messages. Must be thread-safe.
    Store(ctx context.Context, messages []domain.LogMessage) error
    // Close releases resources (e.g., closes database connections).
    Close() error
}
```

**Example:** Let's create an adapter to write logs to a file.

1.  Create a new directory: `internal/adapters/output/file/`
2.  Create a `storage.go` file inside it:

```go
// file: internal/adapters/output/file/storage.go
package file

import (
	"context"
	"encoding/json"
	"os"
	"sync"

	"github.com/shanth1/loggate/internal/core/domain"
)

// Storage implements ports.LogStorage for writing to a file.
type Storage struct {
	file *os.File
	mu   sync.Mutex
}

// New creates a new instance of File Storage.
// path is the DSN from the config.
func New(path string) (*Storage, error) {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	return &Storage{file: f}, nil
}

func (s *Storage) Store(_ context.Context, messages []domain.LogMessage) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, msg := range messages {
		raw, err := json.Marshal(msg)
		if err != nil {
			// In a real application, better error handling is needed here
			continue
		}
		if _, err := s.file.Write(append(raw, '\n')); err != nil {
			return err
		}
	}
	return nil
}

func (s *Storage) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.file.Close()
}
```

### Step 2: Integrate the Adapter into `main.go`

Now, you need to "teach" the application how to create your new adapter on startup.

Edit `cmd/loggate/main.go`:

```go
// cmd/loggate/main.go

package main

import (
    // ... other imports
	"github.com/shanth1/loggate/internal/adapters/output/console"
    "github.com/shanth1/loggate/internal/adapters/output/file" // <-- 1. Import your package
    // ...
)

func main() {
    // ... (initialization)

	storages := make(map[string]ports.LogStorage)

	for storageName, storageCfg := range cfg.Storages {
		if !storageCfg.Enabled {
			continue
		}

		var storage ports.LogStorage
        var err error // <-- Add an error variable

		switch storageCfg.Type {
		case "console":
			storage = console.New()
		case "file": // <-- 2. Add a case for your type
            // DSN is now used as the file path
			storage, err = file.New(storageCfg.DSN)
            if err != nil {
                logger.Fatal().Err(err).Str("storage", storageName).Msg("failed to create file storage")
            }
		default:
			logger.Warn().Msg(fmt.Sprintf("storage type '%s' is enabled but not implemented, skipping", storageCfg.Type))
		}

		if storage != nil {
			storages[storageName] = storage
		}
	}
    // ... (rest of the code)
}
```

### Step 3: Update the Configuration

You can now use the new storage type in `config/config.yaml`:

```yaml
storages:
  # ...
  file_output:
    type: 'file'
    enabled: true
    dsn: '/tmp/loggate_output.log' # Path to the log file

routing_rules:
  - match_condition:
      service: 'inventory-service'
    destinations: ['file_output'] # Route logs to the file
```

After these steps, LogGate will be able to receive logs and save them to the specified file.

## 5.3. Using the Log Generator (`loggen`)

The `loggen` utility is a powerful tool for performance testing and verifying routing rules. Its behavior is configured in the `cmd/loggen/config/config.yaml` file.

- `target`: The address of the LogGate service.
- `load`: Load parameters (`workers` - number of concurrent senders, `rps` - requests per second per worker).
- `templates`: Templates for generating diverse logs. `gofakeit` is used to substitute random data into the `{...}` fields.
