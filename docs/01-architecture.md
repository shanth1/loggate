# 1. LogGate System Architecture

[RU](./01-architecture.ru.md)

## 1.1. Concept

LogGate is designed as a middleware layer between log sources (applications) and their storage systems. Its main task is to relieve applications of the responsibility of _where_ and _how_ to deliver logs. Applications simply send structured messages over UDP, and LogGate handles routing, buffering, and delivery.

This allows for:

- **Reduced Load on Applications:** Sending logs via UDP is a non-blocking "fire-and-forget" operation.
- **Centralized Routing Logic:** Changing the log storage (e.g., migrating from Loki to ClickHouse) requires a configuration change in only one place—LogGate.
- **Fault Tolerance:** LogGate can buffer logs if the target storage is temporarily unavailable (feature under development).

## 1.2. Component Interaction Diagram

```

                                      +-------------------------+
                                      |    External Services    |
                                      +-------------------------+
                                                  |
                                                  | (UDP, JSON Logs)
                                                  v

+---------------------------------------------------------------------------------------------------+
| Docker Host |
| |
| +------------------------+ (stdout) +--------------------+ |
| | loggate-service |------------------------>| promtail | |
| | (Go App, UDP:10514) | | (Log Collector) | |
| +------------------------+ +--------------------+ |
| | ^ | (Logs) |
| | (Metrics /metrics) v |
| | | +--------------------+ |
| +----------+------------------------------------>| loki | |
| | | (Log Aggregator) | |
| v +--------------------+ |
| +------------------------+ ^ |
| | prometheus | | (LogQL Queries) |
| | (Metrics Scraper) |<------------------------+ | |
| +------------------------+ | | |
| ^ ^ | | |
| | +-------------------------------------|---------+ |
| | (PromQL) | (Queries) |
| | v |
| +------------------------+ +------------------------+ |
| | grafana | | cadvisor | |
| | (Visualization) | | (Container Metrics) | |
| +------------------------+ +------------------------+ |
| ^ |
| | (Scrape) |
| +---------------------------------------------+
| |
+---------------------------------------------------------------------------------------------------+

```

## 1.3. Component Description

| Service               | Role in the System                                                                                                                                    | Key Files/Ports                        |
| :-------------------- | :---------------------------------------------------------------------------------------------------------------------------------------------------- | :------------------------------------- |
| **`loggate-service`** | **The core of the system.** Receives logs over UDP, routes them according to `config.yaml`, and outputs them to `stdout`. Exports Prometheus metrics. | `10514/udp`, `9100/tcp`, `config.yaml` |
| **`prometheus`**      | A system for collecting, storing, and processing metrics. Scrapes `loggate`, `loki`, `promtail`, and `cadvisor`.                                      | `9090/tcp`, `prometheus.yaml`          |
| **`loki`**            | A log aggregation and storage system. Receives processed logs from `promtail`.                                                                        | `3100/tcp`, `loki-config.yaml`         |
| **`promtail`**        | An agent for collecting logs. Reads `stdout` from the `loggate` container, parses JSON, enriches it with metadata (labels), and sends it to Loki.     | `promtail-config.yaml`                 |
| **`grafana`**         | A visualization platform. Displays metrics from Prometheus and logs from Loki on a pre-configured dashboard.                                          | `3000/tcp`, `grafana/provisioning/*`   |
| **`cadvisor`**        | Collects resource usage metrics (CPU, memory, network) for all Docker containers on the host.                                                         | `8080/tcp`                             |

## 1.4. Log Processing Flow (Data Flow)

1.  **Ingestion:** An external service sends a JSON message via UDP to port `10514`.
2.  **Decoding:** `loggate-service` (`udp/listener.go`) receives the packet, validates it, and decodes it into a `domain.LogMessage` struct. Required fields: `app` and `service`.
3.  **Routing:** The service core (`service/service.go`) checks the `routing_rules` in `config.yaml`. Based on the `service` and/or `level` fields of the log, a list of destination storages is determined. If no rule matches, the `default_destinations` are used.
4.  **Buffering:** The log is placed into an internal Go channel corresponding to each assigned storage. This prevents blocking the input stream.
5.  **Batch Processing:** A worker for each storage accumulates logs from the channel into a batch. The batch is sent when `batch_size` is reached or `batch_timeout_ms` expires.
6.  **Output:** In the current configuration, the only active storage is `console`. It serializes each log from the batch back into JSON and writes it to the container's `stdout`.
7.  **Collection by Promtail:** `promtail` is configured to read the `stdout` of the `loggate-service` container.
8.  **Enrichment and Indexing:** Using `pipeline_stages` in `promtail-config.yaml`, Promtail parses the JSON string, extracts the `app`, `service`, and `level` fields, and converts them into **Loki labels**. This is a crucial step that makes logs indexable and allows for fast filtering in Grafana.
9.  **Storage:** Promtail sends the processed logs with labels to Loki.
10. **Visualization:** The user opens a dashboard in Grafana, which executes LogQL queries against Loki to display and filter logs.
