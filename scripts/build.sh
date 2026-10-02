#!/usr/bin/env bash
set -euo pipefail
rm -rf dist
mkdir -p dist
cp web/index.html web/style.css web/app.js "$(go env GOROOT)/lib/wasm/wasm_exec.js" dist/
GOOS=js GOARCH=wasm go build -o dist/brotato.wasm ./cmd/wasm
