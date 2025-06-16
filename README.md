# Loggate

1.  **Сборка и первый запуск:**
    ```bash
    docker-compose up --build -d
    ```

2.  **Проверка состояния:**
    ```bash
    docker-compose ps
    ```

3.  **Просмотр логов конкретного сервиса (например, `loggate`):**
    ```bash
    docker-compose logs -f loggate
    ```
    Для `promtail`: `docker-compose logs -f promtail` (убедитесь, что он видит и собирает логи `loggate`).

4.  **Остановка сервисов:**
    ```bash
    docker-compose stop
    ```

5.  **Запуск остановленных сервисов:**
    ```bash
    docker-compose start
    ```

6.  **Остановка и удаление контейнеров:**
    ```bash
    docker-compose down
    ```

7.  **Остановка, удаление контейнеров И УДАЛЕНИЕ VOLUMES (данных Prometheus, Loki, Grafana):**
    ```bash
    docker-compose down -v
    ```
    **ОСТОРОЖНО:** Эта команда удалит все сохраненные метрики, логи и дашборды, если они не были экспортированы.

8.  **Пересборка образа `loggate` и перезапуск:**
    ```bash
    docker-compose up --build -d loggate
    # или если все надо перезапустить с пересборкой loggate
    docker-compose up --build -d
    ```

9.  **Перезагрузка конфигурации Prometheus (если включен `--web.enable-lifecycle`):**
    *   Измените `prometheus/prometheus.yml`.
    *   `curl -X POST http://localhost:9090/-/reload`
    *   Или `docker-compose restart prometheus`.

10. **Доступ к интерфейсам:**
    *   **Prometheus:** `http://localhost:9090`
    *   **Grafana:** `http://localhost:3000` (логин: `admin`, пароль: `YourSecurePassword!`)
    *   **Loki:** API доступен на `http://localhost:3100`, но обычно с ним работают через Grafana.
