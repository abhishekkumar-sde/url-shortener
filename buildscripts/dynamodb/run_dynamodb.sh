#!/bin/bash

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/../default_env.sh"

docker rm -f "$DYNAMODB_CONTAINER_NAME" 2>/dev/null || true

docker run -d \
    --name "$DYNAMODB_CONTAINER_NAME" \
    --network "$DOCKER_NETWORK" \
    -p "$DYNAMODB_EXPOSED_PORT:$DYNAMODB_PORT" \
    amazon/dynamodb-local:latest \
    -jar DynamoDBLocal.jar \
    -sharedDb \
    -inMemory

echo "DynamoDB Local started:"
echo "  Container : $DYNAMODB_CONTAINER_NAME"
echo "  Host      : localhost:$DYNAMODB_EXPOSED_PORT"
echo "  Docker    : $DYNAMODB_CONTAINER_NAME:8000"