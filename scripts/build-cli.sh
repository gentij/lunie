#!/usr/bin/env sh
set -euo pipefail

ROOT_DIR=$(cd "$(dirname "$0")/.." && pwd)
DIST_DIR="$ROOT_DIR/dist"
CLI_DIR="$ROOT_DIR/apps/cli"

mkdir -p "$DIST_DIR"

build() {
  GOOS="$1" GOARCH="$2" OUTPUT_NAME="$3"
  echo "Building $OUTPUT_NAME ($GOOS/$GOARCH)"
  (cd "$CLI_DIR" && GOOS="$GOOS" GOARCH="$GOARCH" go build -o "$DIST_DIR/$OUTPUT_NAME" ./cmd/lunie)
}

build darwin arm64 lunie_darwin_arm64
build darwin amd64 lunie_darwin_amd64
build linux amd64 lunie_linux_amd64
build linux arm64 lunie_linux_arm64
build windows amd64 lunie_windows_amd64.exe

echo "CLI builds complete: $DIST_DIR"
