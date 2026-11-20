#!/usr/bin/env bash
# Respaldo diario de la base de producción con rotación de 7 días.
# Cron sugerido (usuario deploy):  15 3 * * * /opt/plataforma/repo/deploy/scripts/backup.sh production
set -euo pipefail

ENV_NAME="${1:-production}"
BACKUP_DIR="${BACKUP_DIR:-/var/backups/plataforma}"
RETENTION_DAYS="${RETENTION_DAYS:-7}"
CONTAINER="plataforma-${ENV_NAME}-postgres-1"

mkdir -p "$BACKUP_DIR"
chmod 700 "$BACKUP_DIR"

file="$BACKUP_DIR/${ENV_NAME}-$(date +%Y%m%d-%H%M%S).dump"
docker exec "$CONTAINER" sh -c 'pg_dump -U "$POSTGRES_USER" -d "$POSTGRES_DB" -Fc' > "$file"
chmod 600 "$file"

find "$BACKUP_DIR" -name "${ENV_NAME}-*.dump" -mtime "+$RETENTION_DAYS" -delete
echo "Respaldo creado: $file"
