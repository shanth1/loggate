.PHONY: help up down restart logs status shell backup clean destroy test-log

help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@echo "  up             - Start all services in detached mode"
	@echo "  down           - Stop and remove all services"
	@echo "  restart        - Restart all services"
	@echo "  logs           - Follow logs of the loggate service"
	@echo "  status         - Show status of all services"
	@echo "  shell          - Get a shell inside the loggate container"
	@echo "  backup         - Create a cold backup of Loki data"
	@echo "  clean          - Stop services and remove volumes (DELETES ALL DATA)"
	@echo "  destroy        - Alias for clean"
	@echo "  test-log       - Send a single test log message via netcat"
	@echo "  gen-logs       - Run the Go-based log generator"
	@echo "  change-password - Guide to change Grafana admin password"

up:
	@echo "Starting up services..."
	@cp -n .env.example .env || true
	docker-compose up --build -d

down:
	@echo "Stopping services..."
	docker-compose down

restart:
	@echo "Restarting services..."
	docker-compose restart

logs:
	@echo "Following loggate service logs..."
	docker-compose logs -f loggate

status:
	docker-compose ps

shell:
	docker-compose exec loggate-service /bin/sh

backup:
	@echo "Creating Loki data backup..."
	@LOKI_VOLUME=$$(docker volume inspect loggate_loki_data -f '{{.Mountpoint}}')
	@if [ -z "$$LOKI_VOLUME" ]; then echo "Loki data volume not found!"; exit 1; fi
	@BACKUP_FILE="loki-backup-$$(date +%Y-%m-%d_%H-%M-%S).tar.gz"
	@echo "Stopping Loki to ensure data consistency..."
	@docker-compose stop loki promtail
	@echo "Creating archive $$BACKUP_FILE from $$LOKI_VOLUME..."
	@tar -czvf "$$BACKUP_FILE" -C "$$LOKI_VOLUME" .
	@echo "Starting Loki back up..."
	@docker-compose start loki promtail
	@echo "Backup complete: $$BACKUP_FILE"

clean destroy:
	@echo "WARNING: This will permanently delete all logs, metrics, and dashboards."
	@read -p "Are you sure? [y/N] " -r; \
    if [[ $$REPLY =~ ^[Yy]$$ ]]; then \
        echo "Stopping services and removing all data..."; \
        docker-compose down -v; \
    else \
        echo "Operation cancelled."; \
    fi


test-log:
	@echo "Sending a reliably formatted test log..."
	@TIMESTAMP=$$(date -u +'%Y-%m-%dT%H:%M:%SZ'); \
	printf '{"timestamp":"%s","app":"manual-app","service":"manual-test","level":"info","message":"hello from make"}\n' "$$TIMESTAMP" | nc -u -w0 127.0.0.1 10514
	@echo "Test log sent successfully."

gen-logs:
	go run ./cmd/loggen/main.go

gate:
	go run cmd/loggate/main.go -config config/config.yaml


change-password:
	@echo "To change the Grafana admin password:"
	@echo "1. Edit the .env file and set a new GRAFANA_ADMIN_PASSWORD."
	@echo "2. Run 'make down' if the stack is running."
	@echo "3. Run 'make up'. Grafana will start with the new password."
	@echo "NOTE: This only works for the first startup or after deleting the grafana volume."
	@echo "If you want to change it on a running instance, you need to use grafana-cli."
