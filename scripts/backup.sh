#!/bin/sh

set -e

echo "--- LogGate Backup Script ---"
echo ""

echo "[1/6] Checking for 'jq' utility..."
if ! command -v jq &> /dev/null; then
    echo "❌ ERROR: 'jq' is not installed but is required for this script."
    echo "   Please install it to continue (e.g., 'sudo apt-get install jq' or 'brew install jq')."
    exit 1
fi
echo "✅ 'jq' is installed."
echo ""

echo "[2/6] Finding Loki data volume name..."
LOKI_VOLUME_NAME=$(docker inspect loki | jq -r '.[0].Mounts[] | select(.Destination == "/loki") | .Name')

if [ -z "$LOKI_VOLUME_NAME" ]; then
    echo "❌ ERROR: Could not find the Loki data volume. Is the 'loki' container running?"
    exit 1
fi
echo "✅ Loki data volume found: $LOKI_VOLUME_NAME"
echo ""

BACKUP_DIR="backups"
echo "[3/6] Ensuring backup directory '$BACKUP_DIR' exists..."
mkdir -p "$BACKUP_DIR"
echo "✅ Backup directory is ready."
echo ""

BACKUP_FILE="${BACKUP_DIR}/loki-backup-$(date +%Y-%m-%d_%H-%M-%S).tar.gz"

echo "[4/6] Stopping services to ensure data consistency..."
docker-compose stop loki promtail
echo "✅ Services stopped."
echo ""

echo "[5/6] Creating archive '$BACKUP_FILE' using a temporary container..."
docker run --rm \
  -v "${LOKI_VOLUME_NAME}":/data:ro \
  -v "$(pwd)":/backup \
  alpine \
  tar -czvf "/backup/${BACKUP_FILE}" -C /data .

echo "✅ Archive created."
echo ""

echo "[6/6] Restarting services..."
docker-compose start loki promtail
echo "✅ Services restarted."
echo ""

echo "🎉 Backup complete: $BACKUP_FILE"
