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
echo "Starting URL Encoder Server"
echo "======================================"

echo "Project Root : $PROJECT_ROOT"
echo "Dockerfile   : $DOCKERFILE"

# --------------------------------------------------
# Remove existing encoder container
# --------------------------------------------------

echo ""
echo "Removing existing encoder container..."

docker rm -f "$ENCODER_CONTAINER_NAME" 2>/dev/null || true

# --------------------------------------------------
# Build Docker image
# --------------------------------------------------

echo ""
echo "Building URL Encoder Server image..."

docker build \
    --build-arg SERVICE=encoder \
    -f "$DOCKERFILE" \
    -t "${CONTAINER_NAME}-encoder:latest" \
    "$PROJECT_ROOT"

# --------------------------------------------------
# Start encoder
# --------------------------------------------------

echo ""
echo "Starting URL Encoder server..."

docker run -d \
    --name "$ENCODER_CONTAINER_NAME" \
    --network "$DOCKER_NETWORK" \
    -e HTTP_PORT="$ENCODER_PORT" \
    -e REDIS_HOST="${REDIS_CONTAINER_NAME}" \
    -e REDIS_PORT="${REDIS_PORT}" \
    -e DYNAMODB_HOST="${DYNAMODB_CONTAINER_NAME}" \
    -e DYNAMODB_PORT="${DYNAMODB_PORT}" \
    -e DYNAMODB_REGION="local" \
    -e DYNAMODB_TABLE="URLMappings" \
    -e BASE_URL="http://localhost:${LOADBALANCER_EXPOSED_PORT}" \
    -e AWS_ACCESS_KEY_ID="local" \
    -e AWS_SECRET_ACCESS_KEY="local" \
    "${CONTAINER_NAME}-encoder:latest"

echo ""
echo "======================================"
echo "URL Encoder started"
echo "======================================"
echo "Container : $ENCODER_CONTAINER_NAME"
echo "Port      : $ENCODER_PORT"
echo "======================================"