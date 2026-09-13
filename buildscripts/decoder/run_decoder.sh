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
echo "Starting URL Decoder Server"
echo "======================================"

echo "Project Root : $PROJECT_ROOT"
echo "Dockerfile   : $DOCKERFILE"

# --------------------------------------------------
# Remove existing decoder container
# --------------------------------------------------

echo ""
echo "Removing existing decoder container..."

docker rm -f "$DECODER_CONTAINER_NAME" 2>/dev/null || true

# --------------------------------------------------
# Build Docker image
# --------------------------------------------------

echo ""
echo "Building URL Decoder Server image..."

docker build \
    --build-arg SERVICE=decoder \
    -f "$DOCKERFILE" \
    -t "${CONTAINER_NAME}-decoder:latest" \
    "$PROJECT_ROOT"

# --------------------------------------------------
# Start decoder
# --------------------------------------------------

echo ""
echo "Starting URL Decoder server..."

docker run -d \
    --name "$DECODER_CONTAINER_NAME" \
    --network "$DOCKER_NETWORK" \
    -e HTTP_PORT="$DECODER_PORT" \
    -e REDIS_HOST="${REDIS_CONTAINER_NAME}" \
    -e REDIS_PORT="${REDIS_PORT}" \
    -e DYNAMODB_HOST="${DYNAMODB_CONTAINER_NAME}" \
    -e DYNAMODB_PORT="${DYNAMODB_PORT}" \
    -e DYNAMODB_REGION="local" \
    -e DYNAMODB_TABLE="URLMappings" \
    -e AWS_ACCESS_KEY_ID="local" \
    -e AWS_SECRET_ACCESS_KEY="local" \
    "${CONTAINER_NAME}-decoder:latest"

echo ""
echo "======================================"
echo "URL Decoder started"
echo "======================================"
echo "Container : $DECODER_CONTAINER_NAME"
echo "Port      : $DECODER_PORT"
echo "======================================"