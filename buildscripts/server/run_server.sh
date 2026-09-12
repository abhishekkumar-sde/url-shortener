#!/bin/bash

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Project root
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

# Dockerfile
DOCKERFILE="$PROJECT_ROOT/buildscripts/build/Dockerfile"

# Shared environment
source "$PROJECT_ROOT/buildscripts/default_env.sh"

echo "======================================"
echo "Starting URL Shortener"
echo "======================================"

echo "Project Root : $PROJECT_ROOT"
echo "Dockerfile   : $DOCKERFILE"

# --------------------------------------------------
# Remove existing server container
# --------------------------------------------------

echo ""
echo "Removing existing server container..."

docker rm -f "$SERVER_CONTAINER_NAME" 2>/dev/null || true

# --------------------------------------------------
# Build Docker image
# --------------------------------------------------

echo ""
echo "Building URL Shortener image..."

docker build \
    -f "$DOCKERFILE" \
    -t "${CONTAINER_NAME}:latest" \
    "$PROJECT_ROOT"

# --------------------------------------------------
# Start server
# --------------------------------------------------

echo ""
echo "Starting URL Shortener server..."

docker run -d \
    --name "$SERVER_CONTAINER_NAME" \
    --network "$DOCKER_NETWORK" \
    -p "$SERVER_EXPOSED_PORT:$SERVER_PORT" \
    -e HTTP_PORT=8080 \
    -e REDIS_HOST="${REDIS_CONTAINER_NAME}" \
    -e REDIS_PORT="${REDIS_PORT}" \
    -e DYNAMODB_HOST="${DYNAMODB_CONTAINER_NAME}" \
    -e DYNAMODB_PORT="${DYNAMODB_PORT}" \
    -e DYNAMODB_REGION="local" \
    -e DYNAMODB_TABLE="URLMappings" \
    -e BASE_URL="http://localhost:${SERVER_EXPOSED_PORT}" \
    -e AWS_ACCESS_KEY_ID="local" \
    -e AWS_SECRET_ACCESS_KEY="local" \
    "${CONTAINER_NAME}:latest"

echo ""
echo "======================================"
echo "URL Shortener started"
echo "======================================"
echo "Container : $SERVER_CONTAINER_NAME"
echo "API       : http://localhost:$SERVER_EXPOSED_PORT"
echo "======================================"