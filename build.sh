#!/bin/bash

GIT_COMMIT=$(git rev-parse --short HEAD)
BUILD_TIME=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
VERSION=$(git describe --tags --always --dirty)

echo "GIT_COMMIT:$GIT_COMMIT"
echo "BUILD_TIME:$BUILD_TIME"
echo "VERSION:$VERSION"

docker compose build --build-arg VERSION="$VERSION" --build-arg GIT_COMMIT="$GIT_COMMIT" --build-arg BUILD_TIME="$BUILD_TIME"
