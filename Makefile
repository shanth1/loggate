.PHONY: help up down build restart logs status shell backup clean destroy test-log gen-logs change-password config-dev config-prod

help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Service Management:"
	@echo "  up             - Start all services in detached mode. Creates .env if missing."
	@echo "  down           - Stop and remove all service containers."
	@echo "  build          - Force rebuild of the loggate service image."
	@echo "  restart        - Restart all services."
	@echo "  logs           - Follow logs of the loggate service."
	@echo "  status         - Show the status of all services."
	@echo ""
	@echo "Data & Configuration:"
	@echo "  backup         - Create a cold backup of Loki data by running the backup script (requires 'jq')."
	@echo "  restore        - Restore Loki data from a backup file. Usage: make restore BACKUP_FILE=<file.tar.gz>"
	@echo "  clean          - Stop services and REMOVE ALL DATA (volumes). Use with caution."
	@echo "  config-dev     - Print instructions for setting up the .env file for development."
	@echo "  config-prod    - Print instructions for setting up the .env file for production."
	@echo ""
	@echo "Utilities:"
	@echo "  shell          - Get a shell inside the loggate container."
	@echo "  test-log       - Send a single, reliably formatted test log."
	@echo "  gen-logs       - Run the Go-based continuous log generator."
	@echo "  change-password- Guide to changing the Grafana admin password."


# --- Service Management ---

up:
	@echo "Starting up services..."
	@echo "Ensuring .env file exists..."
	@cp -n .env.example .env || true
	docker-compose up -d

down:
	@echo "Stopping services..."
	docker-compose down

build:
	@echo "Forcing a rebuild of the loggate service image..."
	docker-compose build loggate

restart:
	@echo "Restarting services..."
	docker-compose restart

logs:
	@echo "Following loggate service logs..."
	docker-compose logs -f loggate-service

status:
	docker-compose ps


# --- Data & Configuration ---

backup:
	@./scripts/backup.sh

restore:
	@# This block defines the target backup file.
	@# If BACKUP_FILE is passed by the user, it uses that.
	@# Otherwise, it finds the latest .tar.gz file in the 'backups' directory.
	@TARGET_BACKUP="$(BACKUP_FILE)"; \
	if [ -z "$$TARGET_BACKUP" ]; then \
		echo "INFO: BACKUP_FILE not specified. Searching for the latest backup in 'backups/'..."; \
		TARGET_BACKUP=$$(ls -t backups/loki-backup-*.tar.gz 2>/dev/null | head -n 1); \
		if [ -z "$$TARGET_BACKUP" ]; then \
			echo "❌ ERROR: No backup files found in the 'backups/' directory."; \
			echo "   Please create a backup first using 'make backup'."; \
			exit 1; \
		fi; \
		echo "✅ Found latest backup: $$TARGET_BACKUP"; \
	fi; \
	\
	echo "---"; \
	./scripts/restore.sh "$$TARGET_BACKUP"

clean:
	@echo "WARNING: This will permanently delete all logs, metrics, and dashboards."
	@read -p "Are you sure? [y/N] " -r; \
	if [[ $$REPLY =~ ^[Yy]$$ ]]; then \
		echo "Stopping services and removing all data..."; \
		docker-compose down -v; \
	else \
		echo "Operation cancelled."; \
	fi

destroy: clean

env:
	@cp -n .env.example .env || true

config-dev:
	@echo "--- Development Configuration ---"
	@echo "For local development, edit your .env file and set:"
	@echo "DOCKER_HOST_BIND_IP=0.0.0.0"
	@echo "Make sure all resource limits are commented out."
	@echo "Then run 'make up'."

config-prod:
	@echo "--- Production Configuration ---"
	@echo "For production, edit your .env file:"
	@echo "1. Set a strong and unique GRAFANA_ADMIN_PASSWORD."
	@echo "2. Set DOCKER_HOST_BIND_IP=127.0.0.1 for security."
	@echo "3. Uncomment and adjust the CPU/Memory limits to fit your server."
	@echo "Then run 'make up'."


# --- Utilities ---

shell:
	docker-compose exec loggate-service /bin/sh

test-log:
	@echo "Sending a reliably formatted test log..."
	@TIMESTAMP=$$(date -u +'%Y-%m-%dT%H:%M:%SZ'); \
	printf '{"timestamp":"%s","app":"manual-app","service":"manual-test","level":"info","message":"hello from make"}\n' "$$TIMESTAMP" | nc -u -w0 127.0.0.1 10514
	@echo "Test log sent successfully."

gen-logs:
	go run ./cmd/loggen/main.go

change-password:
	@echo "To change the Grafana admin password:"
	@echo "--- For Development (resets all Grafana data) ---"
	@echo "1. Edit the .env file and set a new GRAFANA_ADMIN_PASSWORD."
	@echo "2. Run 'make clean'."
	@echo "3. Run 'make up'. Grafana will start fresh with the new password."
	@echo ""
	@echo "--- For Production (preserves data) ---"
	@echo "1. Choose a new secure password."
	@echo "2. Run the following command, replacing 'new_password' with your choice:"
	@echo "   docker-compose exec grafana grafana-cli admin reset-admin-password 'new_password'"
