#!/bin/bash

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Load shared environment
source "$SCRIPT_DIR/default_env.sh"

echo "======================================"
echo "Starting URL Shortener stack"
echo "======================================"

# --------------------------------------------------
# 1. Create Docker network
# --------------------------------------------------

echo ""
echo "[1/4] Setting up Docker network..."

if docker network inspect "$DOCKER_NETWORK" >/dev/null 2>&1; then
    echo "Docker network already exists: $DOCKER_NETWORK"
else
    echo "Creating Docker network: $DOCKER_NETWORK"
    docker network create "$DOCKER_NETWORK"
fi

# --------------------------------------------------
# 2. Setup Redis
# --------------------------------------------------

echo ""
echo "[2/4] Setting up Redis..."

"$SCRIPT_DIR/redis/run_redis.sh"

# --------------------------------------------------
# 3. Setup DynamoDB
# --------------------------------------------------

echo ""
echo "[3/4] Setting up DynamoDB..."

"$SCRIPT_DIR/dynamodb/run_dynamodb.sh"

# --------------------------------------------------
# 4. Setup URL Shortener server
# --------------------------------------------------

echo ""
echo "[4/4] Setting up URL Shortener server..."

"$SCRIPT_DIR/server/run_server.sh"

# --------------------------------------------------
# Done
# --------------------------------------------------

echo ""
echo "======================================"
echo "URL Shortener stack started"
echo "======================================"
echo ""
echo "API       : http://localhost:${SERVER_EXPOSED_PORT}"
echo "Redis     : localhost:${REDIS_EXPOSED_PORT}"
echo "DynamoDB  : localhost:${DYNAMODB_EXPOSED_PORT}"
echo ""
echo "Docker network: ${DOCKER_NETWORK}"
echo "======================================"