#!/bin/sh
set -eu

ROOT="$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)"
mkdir -p "$ROOT/rootfs/usr/bin"

cd "$ROOT/src"

CGO_ENABLED=0 \
GOOS=linux \
GOARCH=arm64 \
go build \
 -trimpath \
 -ldflags='-s -w' \
 -o "$ROOT/rootfs/usr/bin/torrent-parser" \
.
