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
      # 2. Установить временную метку из поля 'timestamp' в логе
      - timestamp:
          source: timestamp
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
