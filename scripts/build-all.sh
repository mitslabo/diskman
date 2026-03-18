#!/bin/bash

ARCHS="amd64 arm64 arm"

for arch in $ARCHS; do
  echo "Building for $arch..."
  CGO_ENABLED=0 GOOS=linux GOARCH=$arch go build -o dist/diskman.$arch -ldflags="-s -w" -trimpath -tags "static $arch" .
done