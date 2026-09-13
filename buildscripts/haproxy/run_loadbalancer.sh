#!/bin/bash

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Project root
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

# HAProxy config
HAPROXY_CONFIG="$PROJECT_ROOT/buildscripts/haproxy/haproxy.cfg"

# Shared environment
source "$PROJECT_ROOT/buildscripts/default_env.sh"

echo "======================================"
echo "Starting HAProxy"
echo "======================================"

echo "Project Root : $PROJECT_ROOT"
echo "Config       : $HAPROXY_CONFIG"

# --------------------------------------------------
# Remove existing HAProxy container
# --------------------------------------------------

echo ""
echo "Removing existing HAProxy container..."

docker rm -f "$LOADBALANCER_CONTAINER_NAME" 2>/dev/null || true

# --------------------------------------------------
# Start HAProxy
# --------------------------------------------------

echo ""
echo "Starting HAProxy..."

docker run -d \
    --name "$LOADBALANCER_CONTAINER_NAME" \
    --network "$DOCKER_NETWORK" \
    -p "$LOADBALANCER_EXPOSED_PORT:$LOADBALANCER_PORT" \
    -v "$HAPROXY_CONFIG:/usr/local/etc/haproxy/haproxy.cfg:ro" \
    haproxy:3.0-alpine

echo ""
echo "======================================"
echo "HAProxy started"
echo "======================================"
echo "Container : $LOADBALANCER_CONTAINER_NAME"
echo "API       : http://localhost:$LOADBALANCER_EXPOSED_PORT"
echo "======================================"