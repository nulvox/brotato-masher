#!/usr/bin/env bash
set -euo pipefail
root="$(cd "$(dirname "$0")/.." && pwd)"
server_pid=''
tmpdir="$(mktemp -d)"
cleanup() { [[ -n "$server_pid" ]] && kill "$server_pid" 2>/dev/null || true; rm -rf "$tmpdir"; }
trap cleanup EXIT
(go run ./tests/static_server.go -root "$root/dist" >"$tmpdir/server.log" 2>&1 & echo $! >"$tmpdir/server.pid")
server_pid="$(cat "$tmpdir/server.pid")"
for _ in {1..30}; do curl -fsS http://127.0.0.1:8765/ >/dev/null 2>&1 && break; sleep .2; done
chromium --headless --no-sandbox --disable-gpu --dump-dom --virtual-time-budget=2500 http://127.0.0.1:8765/ >"$tmpdir/browser-dom"
grep -q 'Brotato Save Editor' "$tmpdir/browser-dom"
grep -q 'Load a save' "$tmpdir/browser-dom"
printf 'chromium browser smoke passed\n'
