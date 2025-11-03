# 3. Configuration

[RU](./03-configuration.ru.md)

The system is configured via several key files. The core logic of LogGate is controlled by `config.yaml`, while deployment parameters are defined in the `.env` file.

## 3.1. Main LogGate Configuration (`config/config.yaml`)

This file defines the behavior of the `loggate` service.

```yaml
# config/config.yaml

server:
  log_address: ':10514' # Address and UDP port for receiving logs
  info_address: ':9100' # Address and TCP port for Prometheus metrics (/metrics)

performance:
  buffer_size: 1000 # Size of the internal buffer (Go channel) for each storage
  batch_size: 1000 # Max number of logs in a single batch to be sent to a storage
  batch_timeout_ms: 5000 # Timeout in ms, after which a batch is sent even if it's not full

storages:
  # Storage name. Used in routing rules.
  console_debug:
    type: 'console' # Adapter type. Must correspond to an implemented adapter.
    enabled: true # Whether this storage is enabled.

  clickhouse_main:
    type: 'clickhouse'
    enabled: false
    # Connection string. It is recommended to use environment variables.
    dsn: 'clickhouse://<user>:<password>@<host>:<port>/<database>?[params]'

routing_rules:
  # Rule 1: Logs from auth-service with level INFO
  - match_condition:
      service: 'auth-service'
      level: 'INFO'
    destinations: ['console_debug'] # Send only to console_debug

  # Rule 2: All logs from payment-gateway (any level)
  - match_condition:
      service: 'payment-gateway'
    destinations: ['console_debug'] # Send to console_debug

# List of storages where logs are sent if they don't match any rule
default_destinations: ['console_debug']
```

### Details of `config.yaml` sections:

- **`performance`**: Manages the batch sending mechanism.
  - `buffer_size`: How many logs can "wait" in the queue before new logs start being dropped. Protects against short-term load spikes.
  - `batch_size` / `batch_timeout_ms`: A batch is sent when _either_ of these conditions is met. This strikes a balance between latency and efficiency.

- **`storages`**: Defines "where" logs can be sent. Each entry is an instance of an output adapter. The name (`console_debug`, `clickhouse_main`) is a unique identifier used in `destinations`.

- **`routing_rules`**: The heart of the routing logic.
  - Rules are checked sequentially. The **first matching** rule is applied.
  - `match_condition` can contain one or more fields. A log must match **all** fields in the condition.
  - If a field is omitted from `match_condition` (e.g., `level`), it is ignored during the check.

## 3.2. Environment Variables (`.env`)

This file is used by `docker-compose` to substitute values into `docker-compose.yaml`.

| Variable                               | Description                                                                                                  | Default              |
| :------------------------------------- | :----------------------------------------------------------------------------------------------------------- | :------------------- |
| `GRAFANA_ADMIN_USER`                   | Grafana admin username.                                                                                      | `admin`              |
| `GRAFANA_ADMIN_PASSWORD`               | Grafana admin password. **Must be changed for production!**                                                  | `admin`              |
| `DOCKER_HOST_BIND_IP`                  | The host IP address on which ports will be published. `0.0.0.0` for development, `127.0.0.1` for production. | `0.0.0.0`            |
| `LOGGATE_UDP_PORT`                     | External UDP port for LogGate.                                                                               | `10514`              |
| `..._PORT`                             | External TCP ports for other services.                                                                       | (see `.env.example`) |
| `..._LIMIT_CPUS` / `..._LIMIT_MEM`     | **Hard limits** on container resources. The container will not be able to consume more.                      | `0` (none)           |
| `..._RESERVE_CPUS` / `..._RESERVE_MEM` | **Reserved** resources for the container.                                                                    | `0` (none)           |

## 3.3. Promtail Configuration (`promtail/promtail-config.yaml`)

This file is critical for the correct indexing of logs in Loki.

```yaml
# ... (server, positions, clients sections) ...

scrape_configs:
  - job_name: docker
    docker_sd_configs:
      - host: unix:///var/run/docker.sock
        refresh_interval: 5s
        filters:
          # Collect logs only from the loggate-service container
          - name: name
            values: ['loggate-service']
    relabel_configs:
      - source_labels: ['__meta_docker_container_name']
        regex: '/(.*)'
        target_label: 'container'

    # The most important part: the log processing pipeline
    pipeline_stages:
      # 1. Parse the log line as JSON
      - json:
          expressions:
            app: app
            level: level
            service: service
            trace_id: fields.trace_id # Example of extracting a nested field
      # 2. Set the time from the 'time' field in the log
      - time:
          source: time
          format: RFC3339
      # 3. Create Loki labels from the extracted fields
      - labels:
          app:
          level:
          service:
      # 4. Define what will be the message body
      - output:
          source: 'log' # The original field containing the entire JSON string
```

**Key Point:** Fields converted into `labels` (`app`, `level`, `service`) become indexed in Loki. This allows for high-performance queries and filtering in Grafana. All other data remains in the message body and is available for full-text search.

================================================
FILE: docs/03-configuration.ru.md
================================================
[EN](./03-configuration.md)

# 3. Конфигурация

Система настраивается через несколько ключевых файлов. Основная логика LogGate управляется файлом `config.yaml`, а параметры развертывания — файлом `.env`.

## 3.1. Основная конфигурация LogGate (`config/config.yaml`)

Этот файл определяет поведение сервиса `loggate`.

```yaml
# config/config.yaml

server:
  log_address: ':10514' # Адрес и UDP-порт для приема логов
  info_address: ':9100' # Адрес и TCP-порт для метрик Prometheus (/metrics)

performance:
  buffer_size: 1000 # Размер внутреннего буфера (Go-канала) для каждого хранилища
  batch_size: 1000 # Макс. кол-во логов в одном пакете для отправки в хранилище
  batch_timeout_ms: 5000 # Таймаут в мс, после которого пакет отправляется, даже если не полон

storages:
  # Имя хранилища. Используется в правилах маршрутизации.
  console_debug:
    type: 'console' # Тип адаптера. Должен соответствовать реализованному адаптеру.
    enabled: true # Включено ли данное хранилище.

  clickhouse_main:
    type: 'clickhouse'
    enabled: false
    # Строка подключения. Рекомендуется использовать переменные окружения.
    dsn: 'clickhouse://<user>:<password>@<host>:<port>/<database>?[params]'

routing_rules:
  # Правило 1: Логи от auth-service с уровнем INFO
  - match_condition:
      service: 'auth-service'
      level: 'INFO'
    destinations: ['console_debug'] # Отправлять только в console_debug

  # Правило 2: Все логи от payment-gateway (любой уровень)
  - match_condition:
      service: 'payment-gateway'
    destinations: ['console_debug'] # Отправлять в console_debug

# Список хранилищ, куда отправляются логи, если они не подошли ни под одно правило
default_destinations: ['console_debug']
```

### Детали секций `config.yaml`:

- **`performance`**: Управляет механикой пакетной отправки.
  - `buffer_size`: Сколько логов может "ожидать" в очереди, прежде чем новые логи начнут отбрасываться. Защищает от кратковременных всплесков нагрузки.
  - `batch_size` / `batch_timeout_ms`: Пакет отправляется, когда выполняется _любое_ из этих условий. Это баланс между задержкой и эффективностью.

- **`storages`**: Определяет "куда" можно отправлять логи. Каждая запись — это инстанс адаптера вывода. Имя (`console_debug`, `clickhouse_main`) является уникальным идентификатором, который используется в `destinations`.

- **`routing_rules`**: Сердце маршрутизации.
  - Правила проверяются последовательно. Применяется **первое совпавшее** правило.
  - `match_condition` может содержать одно или несколько полей. Лог должен соответствовать **всем** полям в условии.
  - Если поле в `match_condition` опущено (например, `level`), оно игнорируется при проверке.

## 3.2. Переменные окружения (`.env`)

Этот файл используется `docker-compose` для подстановки значений в `docker-compose.yaml`.

| Переменная                             | Описание                                                                                                   | По умолчанию         |
| :------------------------------------- | :--------------------------------------------------------------------------------------------------------- | :------------------- |
| `GRAFANA_ADMIN_USER`                   | Логин администратора Grafana.                                                                              | `admin`              |
| `GRAFANA_ADMIN_PASSWORD`               | Пароль администратора Grafana. **Обязательно смените для production!**                                     | `admin`              |
| `DOCKER_HOST_BIND_IP`                  | IP-адрес хоста, на котором будут опубликованы порты. `0.0.0.0` для разработки, `127.0.0.1` для production. | `0.0.0.0`            |
| `LOGGATE_UDP_PORT`                     | Внешний UDP-порт для LogGate.                                                                              | `10514`              |
| `..._PORT`                             | Внешние TCP-порты для других сервисов.                                                                     | (см. `.env.example`) |
| `..._LIMIT_CPUS` / `..._LIMIT_MEM`     | **Жесткие лимиты** на ресурсы контейнера. Контейнер не сможет потребить больше.                            | `0` (нет)            |
| `..._RESERVE_CPUS` / `..._RESERVE_MEM` | **Гарантированно выделенные** ресурсы для контейнера.                                                      | `0` (нет)            |

## 3.3. Конфигурация Promtail (`promtail/promtail-config.yaml`)

Этот файл критически важен для правильной индексации логов в Loki.

```yaml
# ... (server, positions, clients sections) ...

scrape_configs:
  - job_name: docker
    docker_sd_configs:
      - host: unix:///var/run/docker.sock
        refresh_interval: 5s
        filters:
          # Собираем логи только с контейнера loggate-service
          - name: name
            values: ['loggate-service']
    relabel_configs:
      - source_labels: ['__meta_docker_container_name']
        regex: '/(.*)'
        target_label: 'container'

    # Самая важная часть: конвейер обработки логов
    pipeline_stages:
      # 1. Распарсить строку лога как JSON
      - json:
          expressions:
            app: app
            level: level
            service: service
            trace_id: fields.trace_id # Пример извлечения вложенного поля
      # 2. Установить временную метку из поля 'time' в логе
      - time:
          source: time
          format: RFC3339
      # 3. Создать метки (labels) Loki из извлеченных полей
      - labels:
          app:
          level:
          service:
      # 4. Определить, что будет являться телом сообщения
      - output:
          source: 'log' # Изначальное поле, содержащее всю JSON строку
```

**Ключевой момент:** Поля, превращенные в `labels` (`app`, `level`, `service`), становятся индексируемыми в Loki. Это позволяет выполнять высокопроизводительные запросы и фильтрацию в Grafana. Все остальные данные остаются в теле сообщения и доступны для полнотекстового поиска.
