#!/bin/sh

# Exit immediately if a command exits with a non-zero status.
set -e

echo "--- LogGate Restore Script ---"
echo ""

# 1. Check if a backup file was provided as an argument
if [ -z "$1" ]; then
    echo "❌ ERROR: You must specify a backup file to restore."
    echo "   Usage example: $0 backups/loki-backup-YYYY-MM-DD_HH-MM-SS.tar.gz"
    exit 1
fi
BACKUP_FILE=$1

# Check if the specified file exists
if [ ! -f "$BACKUP_FILE" ]; then
    echo "❌ ERROR: Backup file not found: '$BACKUP_FILE'"
    exit 1
fi

# 2. Display a warning about the destructive nature of the operation
echo "⚠️  WARNING: THIS IS A DESTRUCTIVE OPERATION! ⚠️"
echo "This script will:"
echo "1. Stop the loki and promtail services."
echo "2. COMPLETELY WIPE all current log data in Loki."
echo "3. Restore data from the file '$BACKUP_FILE'."
echo ""
read -p "Are you absolutely sure you want to continue? [y/N] " -r
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "Operation cancelled."
    exit 1
fi
echo ""

# 3. Check for the 'jq' utility
echo "[1/6] Checking for 'jq' utility..."
if ! command -v jq &> /dev/null; then
    echo "❌ ERROR: 'jq' is not installed but is required for this script."
    echo "   Please install it to continue."
    exit 1
fi
echo "✅ 'jq' is installed."
echo ""

# 4. Find the Loki data volume name
echo "[2/6] Finding Loki data volume name..."
LOKI_VOLUME_NAME=$(docker inspect loki | jq -r '.[0].Mounts[] | select(.Destination == "/loki") | .Name')

if [ -z "$LOKI_VOLUME_NAME" ]; then
    echo "❌ ERROR: Could not find the Loki data volume. Is the 'loki' container running?"
    exit 1
fi
echo "✅ Loki data volume found: $LOKI_VOLUME_NAME"
echo ""

# 5. Stop services
echo "[3/6] Stopping services to ensure data consistency..."
docker-compose stop loki promtail
echo "✅ Services stopped."
echo ""

# 6. Clear existing data
echo "[4/6] Clearing existing data in volume '$LOKI_VOLUME_NAME'..."
docker run --rm \
  -v "${LOKI_VOLUME_NAME}":/data \
  alpine \
  sh -c "find /data -mindepth 1 -delete" # <<< THIS IS THE FIX
echo "✅ Volume cleared."
echo ""

# 7. Restore data from the archive
echo "[5/6] Restoring data from '$BACKUP_FILE'..."
docker run --rm \
  -v "${LOKI_VOLUME_NAME}":/data \
  -v "$(pwd)":/backup \
  alpine \
  tar -xzvf "/backup/${BACKUP_FILE}" -C /data
echo "✅ Data restored."
echo ""

# 8. Restart services
echo "[6/6] Restarting services..."
docker-compose start loki promtail
echo "✅ Services restarted."
echo ""

echo "🎉 Restore operation completed successfully!"
