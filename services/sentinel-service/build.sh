#!/bin/bash

# Check if version argument is provided
if [ -z "$1" ]
  then
    echo "Error: Please provide version number (e.g. ./build.sh 1.3)"
    exit 1
fi

VERSION=$1
IMAGE_NAME="himanshudhiman/dashboard-backend"

echo "Building for version v${VERSION}..."

# Build for Linux
GOOS=linux GOARCH=amd64 go build -o main
if [ $? -ne 0 ]; then
    echo "Go build failed"
    exit 1
fi

# Build Docker image
docker build -t ${IMAGE_NAME}:v${VERSION} .
if [ $? -ne 0 ]; then
    echo "Docker build failed"
    exit 1
fi

# Push to Docker Hub
docker push ${IMAGE_NAME}:v${VERSION}
if [ $? -ne 0 ]; then
    echo "Docker push failed"
    exit 1
fi

echo "Successfully built and pushed version v${VERSION}"