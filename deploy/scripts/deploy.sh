#!/usr/bin/env bash
# Despliega una versión en el VPS con rollback automático.
# Uso: deploy.sh <staging|production> <image-tag>
#
# 1. Descarga las imágenes de la versión.
# 2. Aplica migraciones (deben ser compatibles con la versión anterior).
# 3. Actualiza los servicios y espera sus healthchecks.
# 4. Si algo falla, vuelve a la última versión que funcionó.
set -euo pipefail

ENV_NAME="${1:?Uso: deploy.sh <staging|production> <image-tag>}"
IMAGE_TAG="${2:?Falta el tag de la imagen}"

case "$ENV_NAME" in
  staging | production) ;;
  *) echo "Entorno inválido: $ENV_NAME" >&2; exit 2 ;;
esac

DEPLOY_DIR="${DEPLOY_DIR:-/opt/plataforma}"
COMPOSE_FILE="$(cd "$(dirname "$0")/.." && pwd)/docker-compose.yml"
ENV_FILE="$DEPLOY_DIR/$ENV_NAME.env"
LAST_GOOD="$DEPLOY_DIR/$ENV_NAME.last-good"

[[ -f "$ENV_FILE" ]] || { echo "No existe $ENV_FILE" >&2; exit 2; }

export ENV_NAME
compose() {
  IMAGE_TAG="$1" docker compose -f "$COMPOSE_FILE" --env-file "$ENV_FILE" "${@:2}"
}

docker network inspect edge >/dev/null 2>&1 || docker network create edge >/dev/null

rollback() {
  local previous
  previous="$(cat "$LAST_GOOD" 2>/dev/null || true)"
  if [[ -z "$previous" || "$previous" == "$IMAGE_TAG" ]]; then
    echo "::error::El despliegue falló y no hay una versión anterior para restaurar" >&2
    exit 1
  fi
  echo "::warning::El despliegue de $IMAGE_TAG falló; restaurando $previous"
  compose "$previous" up -d --remove-orphans --wait --wait-timeout 120
  exit 1
}

echo "==> Descargando imágenes $IMAGE_TAG para $ENV_NAME"
compose "$IMAGE_TAG" pull --quiet

echo "==> Base de datos y migraciones"
compose "$IMAGE_TAG" up -d --wait postgres
compose "$IMAGE_TAG" run --rm migrate || rollback

echo "==> Actualizando servicios"
compose "$IMAGE_TAG" up -d --remove-orphans --wait --wait-timeout 120 || rollback

echo "$IMAGE_TAG" > "$LAST_GOOD"
docker image prune -f --filter "until=168h" >/dev/null
echo "==> $ENV_NAME desplegado con $IMAGE_TAG"
