#!/bin/sh

set -e

echo "--- LogGate Backup Script ---"
echo ""

echo "[1/5] Checking for 'jq' utility..."
if ! command -v jq &> /dev/null; then
    echo "❌ ERROR: 'jq' is not installed but is required for this script."
    echo "   Please install it to continue:"
    echo "   - On macOS: brew install jq"
    echo "   - On Debian/Ubuntu: sudo apt-get install jq"
    exit 1
fi
echo "✅ 'jq' is installed."
echo ""

echo "[2/5] Finding Loki data volume name..."
LOKI_VOLUME_NAME=$(docker inspect loki | jq -r '.[0].Mounts[] | select(.Destination == "/loki") | .Name')

if [ -z "$LOKI_VOLUME_NAME" ]; then
    echo "❌ ERROR: Could not find the Loki data volume. Is the 'loki' container running?"
    exit 1
fi
echo "✅ Loki data volume found: $LOKI_VOLUME_NAME"
echo ""

BACKUP_FILE="loki-backup-$(date +%Y-%m-%d_%H-%M-%S).tar.gz"

echo "[3/5] Stopping services to ensure data consistency..."
docker-compose stop loki promtail
echo "✅ Services stopped."
echo ""

echo "[4/5] Creating archive '$BACKUP_FILE' using a temporary container..."
docker run --rm \
  -v "${LOKI_VOLUME_NAME}":/data:ro \
  -v "$(pwd)":/backup \
  alpine \
  tar -czvf "/backup/${BACKUP_FILE}" -C /data .

echo "✅ Archive created."
echo ""

echo "[5/5] Restarting services..."
docker-compose start loki promtail
echo "✅ Services restarted."
echo ""

echo "🎉 Backup complete: $BACKUP_FILE"
