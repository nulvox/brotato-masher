# Brotato Save Editor

A private, static Brotato save editor. Save data is processed in the browser: the Go/WASM core validates and edits JSON, while the small JavaScript wrapper handles files and downloads.

## Development

Enter the reproducible Nix environment:

```sh
nix develop
```

Build and validate everything:

```sh
nix develop --command bash -c '
  go test ./...
  bash scripts/build.sh
  bash tests/validate.sh dist
  bash tests/browser-smoke.sh
'
```

Serve locally after building:

```sh
nix develop --command bash -c 'go run ./tests/static_server.go -root dist'
```

Then open <http://127.0.0.1:8765/>. The editor never uploads the selected save and never writes to the source file.

## GitHub Pages

`.github/workflows/pages.yml` tests pull requests and deploys `dist/` to GitHub Pages after pushes to `main`. In the repository settings, set **Pages → Build and deployment → Source** to **GitHub Actions**. If the default branch is not `main`, update the workflow branch filter.

The workflow uses the repository flake for Go, Node, Chromium, Playwright tooling, and `jq`. `tests/browser-smoke.sh` uses Chromium headless as the deterministic baseline so it can run without downloading browser packages at test time.

## Save safety

1. Close Brotato before replacing a save.
2. Keep a backup of the original `save_v3_0.json`.
3. Load the original into the editor.
4. Download `save_v3_0_edited.json`.
5. Replace the game save only after verifying the exported file.

The editor currently displays numeric IDs because the inspected save format does not contain stable human-readable names. It supports unlock/challenge collections, lifetime statistics, purchase and enemy counters, read-announcement arrays, and per-character/per-zone difficulty selection. Existing beaten-wave records are preserved and shown for reference. It does not edit unknown fields or active run state.
