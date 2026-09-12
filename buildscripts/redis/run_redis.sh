#!/bin/bash

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/../default_env.sh"

docker rm -f "$REDIS_CONTAINER_NAME" 2>/dev/null || true

docker run -d \
    --name "$REDIS_CONTAINER_NAME" \
    --network "$DOCKER_NETWORK" \
    -p "$REDIS_EXPOSED_PORT:$REDIS_PORT" \
    redis:7-alpine

echo "Redis started:"
echo "  Container : $REDIS_CONTAINER_NAME"
echo "  Host      : localhost:$REDIS_EXPOSED_PORT"
echo "  Docker    : $REDIS_CONTAINER_NAME:6379"