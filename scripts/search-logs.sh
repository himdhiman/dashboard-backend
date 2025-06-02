#!/bin/bash

# Ask for container name and correlation ID
read -p "Enter Docker container name or ID: " CONTAINER
read -p "Enter Correlation ID to search: " CORRELATION_ID

# Check if container exists
if ! docker ps -a --format '{{.Names}}' | grep -wq "$CONTAINER"; then
    echo "Error: Container '$CONTAINER' not found."
    exit 1
fi

echo "Searching logs in container '$CONTAINER' for correlation ID '$CORRELATION_ID'..."
echo "----------------------------------------------------------"

# Filter logs
docker logs "$CONTAINER" 2>&1 | grep --color=always --line-buffered "$CORRELATION_ID"
