#!/usr/bin/env bash
set -euo pipefail
root="$(cd "$(dirname "$0")/.." && pwd)"
site="${1:-$root/dist}"
[[ -f "$site/index.html" && -f "$site/app.js" && -f "$site/style.css" && -f "$site/brotato.wasm" && -f "$site/wasm_exec.js" ]]
grep -q 'brotato.wasm' "$site/app.js"
grep -q 'wasm_exec.js' "$site/index.html"
! grep -Eq 'https?://|XMLHttpRequest|navigator\.sendBeacon' "$site/app.js"
[[ "$(stat -c %s "$site/brotato.wasm" 2>/dev/null || stat -f %z "$site/brotato.wasm")" -gt 1000 ]]
if command -v jq >/dev/null; then
  jq -e '.version == 3 and (.characters_unlocked|type) == "array" and (.data|type) == "object"' "$root/fixtures/save_v3.json" >/dev/null
fi
printf 'static validation passed: %s\n' "$site"
