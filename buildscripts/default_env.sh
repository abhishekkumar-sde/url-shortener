#!/bin/bash

export CONTAINER_NAME="urlshortener"

# Docker network
export DOCKER_NETWORK="${CONTAINER_NAME}-network"

# Redis
export REDIS_CONTAINER_NAME="${CONTAINER_NAME}-redis"
export REDIS_EXPOSED_PORT="10000"
export REDIS_PORT="6379"

# DynamoDB Local
export DYNAMODB_CONTAINER_NAME="${CONTAINER_NAME}-dynamodb"
export DYNAMODB_EXPOSED_PORT="10001"
export DYNAMODB_PORT="8000"

# URL Shortener server
export SERVER_CONTAINER_NAME="${CONTAINER_NAME}-server"
export SERVER_EXPOSED_PORT="10002"
export SERVER_PORT="8080"