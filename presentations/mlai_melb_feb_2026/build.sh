#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")"

IMAGE="aigogo-slides-build"
CONTAINER="aigogo-slides-tmp"

rm -rf slides
docker rm "$CONTAINER" 2>/dev/null || true

docker build -t "$IMAGE" .
docker create --name "$CONTAINER" "$IMAGE"
docker cp "$CONTAINER":/app/slides ./slides
docker rm "$CONTAINER"
docker rmi "$IMAGE"

echo "Slides built to slides/"
