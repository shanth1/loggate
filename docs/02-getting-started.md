# 2. Installation and Startup

[RU](./02-getting-started.ru.md)

## 2.1. Prerequisites

- **Docker Engine:** Version 20.10.0+
- **Docker Compose:** Version 1.29.0+ (or `docker compose` v2+)
- **GNU Make:** For convenient use of commands from the `Makefile`.
- **Go:** Version 1.23+ (only for using the `loggen` log generator).
- **Git:** For cloning the repository.

## 2.2. Local Launch (for Development)

The steps below deploy the full LogGate stack on your local machine.

### Step 1: Clone the Repository

```bash
git clone <your-repository-url>
cd loggate
```

### Step 2: Create the Environment File

Create a `.env` file from the template. The default values are suitable for local development.

```bash
cp .env.example .env
```

> **Note:** In the `.env` file, `DOCKER_HOST_BIND_IP` is set to `0.0.0.0` by default, which makes the services accessible from all network interfaces on your machine. For production, it is recommended to change this value to `127.0.0.1`.

### Step 3: Build and Run

Use `make` to start all services. The `up` command will automatically build the `loggate` image and run all containers in the background.

```bash
make up
```

### Step 4: Check the Status

Ensure that all containers have started successfully and are in the `Up` or `running` state.

```bash
make status
# or
docker-compose ps
```

### Step 5: Generate Test Data

To test the system, run the built-in log generator. It will send UDP packets to `localhost:10514` according to the configuration in `cmd/loggen/config/config.yaml`.

```bash
make gen-logs
```

You can also send a single test log:

```bash
make test-log
```

### Step 6: Access Web Interfaces

- **Grafana:** [http://localhost:3000](http://localhost:3000)
  - **Login:** `admin`
  - **Password:** `admin` (or the value of `GRAFANA_ADMIN_PASSWORD` from `.env`)
  - The `LogGate Application` dashboard will already be configured and available.
- **Prometheus:** [http://localhost:9090](http://localhost:9090)
  - Check the status of targets for `loggate`, `loki`, etc.
- **cAdvisor:** [http://localhost:8080](http://localhost:8080)
  - View container performance metrics.

## 2.3. Production Deployment

Moving to production requires additional steps to ensure security, stability, and reliability.

1.  **`.env` Configuration:**
    - **MANDATORY:** Set a complex and unique password for `GRAFANA_ADMIN_PASSWORD`.
    - Set `DOCKER_HOST_BIND_IP=127.0.0.1`. This will restrict access to ports (Prometheus, Loki, etc.) to the host machine only. Access to Grafana should be handled through a reverse proxy (Nginx, Traefik) with HTTPS configured.
    - Uncomment and configure resource limits (CPU, Memory) for each service according to your server's capabilities.

2.  **`config/config.yaml` Configuration:**
    - Configure real storages (e.g., `clickhouse_main`), specifying the correct DSNs and credentials. It is recommended to use environment variables for secrets.
    - Disable `console_debug` (`enabled: false`) to avoid unnecessary load on the Docker logging system.

3.  **Data Management (Volumes):**
    - Service data (`prometheus_data`, `loki_data`, `grafana_data`) is stored in Docker volumes on the host machine. Ensure there is enough disk space.
    - Set up regular backups for these volumes. See the [Administration](04-administration.md#42-backup-and-restore) section.

4.  **Startup:**
    - Always rebuild images with the latest changes before deploying to production.
    ```bash
    # Ensure you are in the project root with docker-compose.yaml
    docker-compose build
    docker-compose up -d
    ```
