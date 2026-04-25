#!/bin/bash

# Simple development script for ecosystem-engine

set -e

# Load .env if it exists
if [ -f .env ]; then
    export $(grep -v '^#' .env | xargs)
fi

COMMAND=$1

case $COMMAND in
    "build")
        echo "Building..."
        go build -o api cmd/api/main.go
        ;;
    "run")
        echo "Running..."
        go run cmd/api/main.go
        ;;
    "migrate")
        ACTION=$2
        if [ -z "$ACTION" ]; then ACTION="up"; fi
        echo "Migrating $ACTION..."
        goose -dir migrations postgres "$DATABASE_URL" $ACTION
        ;;
    "lint")
        echo "Linting..."
        golangci-lint run
        ;;
    "help"|*)
        echo "Usage: ./scripts/dev.sh [build|run|migrate|lint|help]"
        ;;
esac
