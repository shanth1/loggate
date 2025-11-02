# 5. Разработка и расширение

Этот раздел предназначен для разработчиков, которые хотят внести изменения в LogGate или расширить его функциональность.

## 5.1. Структура проекта

```

├── cmd/ # Точки входа приложений
│ ├── loggate/ # Основной сервис LogGate
│ └── loggen/ # Утилита для генерации логов
├── config/ # Конфигурация для LogGate по умолчанию
├── docs/ # Документация проекта
├── internal/ # Исходный код приложения
│ ├── adapters/ # Адаптеры (порты и адаптеры)
│ │ ├── input/ # Входные адаптеры (e.g., UDP, TCP)
│ │ └── output/ # Выходные адаптеры (e.g., console, clickhouse)
│ ├── common/ # Общий код (логгер, константы)
│ ├── config/ # Логика загрузки и парсинга конфигурации
│ └── core/ # Ядро бизнес-логики
│ ├── domain/ # Основные структуры данных (e.g., LogMessage)
│ ├── ports/ # Интерфейсы для взаимодействия с ядром
│ └── service/ # Реализация основной бизнес-логики
├── scripts/ # Скрипты для администрирования (backup, restore)
├── Makefile # Утилиты для сборки, запуска и управления
├── go.mod / go.sum # Зависимости Go
└── docker-compose.yaml # Определение стека сервисов

```

## 5.2. Добавление нового хранилища (Output Adapter)

Архитектура LogGate позволяет легко добавлять поддержку новых систем хранения логов. Для этого нужно реализовать адаптер, который удовлетворяет интерфейсу `ports.LogStorage`.

### Шаг 1: Реализация интерфейса `LogStorage`

Интерфейс определен в `internal/core/ports/ports.go`:

```go
type LogStorage interface {
    // Store сохраняет пакет сообщений. Должен быть потокобезопасным.
    Store(ctx context.Context, messages []domain.LogMessage) error
    // Close освобождает ресурсы (например, закрывает соединения с БД).
    Close() error
}
```

**Пример:** Создадим адаптер для записи логов в файл.

1.  Создайте новую директорию: `internal/adapters/output/file/`
2.  Создайте в ней файл `storage.go`:

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

// Storage реализует ports.LogStorage для записи в файл.
type Storage struct {
	file *os.File
	mu   sync.Mutex
}

// New создает новый экземпляр File Storage.
// path - это DSN из конфига.
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
			// В реальном приложении здесь нужна лучшая обработка ошибок
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

### Шаг 2: Интеграция адаптера в `main.go`

Теперь нужно "научить" приложение создавать ваш новый адаптер при старте.

Отредактируйте `cmd/loggate/main.go`:

```go
// cmd/loggate/main.go

package main

import (
    // ... другие импорты
	"github.com/shanth1/loggate/internal/adapters/output/console"
    "github.com/shanth1/loggate/internal/adapters/output/file" // <-- 1. Импортируйте ваш пакет
    // ...
)

func main() {
    // ... (инициализация)

	storages := make(map[string]ports.LogStorage)

	for storageName, storageCfg := range cfg.Storages {
		if !storageCfg.Enabled {
			continue
		}

		var storage ports.LogStorage
        var err error // <-- Добавим переменную для ошибок

		switch storageCfg.Type {
		case "console":
			storage = console.New()
		case "file": // <-- 2. Добавьте case для вашего типа
            // DSN теперь используется как путь к файлу
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
    // ... (дальнейший код)
}
```

### Шаг 3: Обновление конфигурации

Теперь вы можете использовать новый тип хранилища в `config/config.yaml`:

```yaml
storages:
  # ...
  file_output:
    type: 'file'
    enabled: true
    dsn: '/tmp/loggate_output.log' # Путь к файлу логов

routing_rules:
  - match_condition:
      service: 'inventory-service'
    destinations: ['file_output'] # Направляем логи в файл
```

После этих шагов LogGate сможет принимать логи и сохранять их в указанный файл.

## 5.3. Использование генератора логов (`loggen`)

Утилита `loggen` — мощный инструмент для тестирования производительности и проверки правил маршрутизации. Её поведение настраивается в файле `cmd/loggen/config/config.yaml`.

- `target`: Адрес LogGate.
- `load`: Параметры нагрузки (`workers` - количество параллельных отправителей, `rps` - запросов в секунду на одного воркера).
- `templates`: Шаблоны для генерации разнообразных логов. `gofakeit` используется для подстановки случайных данных в поля `{...}`.
