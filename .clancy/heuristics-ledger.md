# Internal Heuristics Ledger
_LOCAL ONLY — no upstream/public reporting._

| id | heuristic | why | provenance | landed-in | standing |
|---|---|---|---|---|---|
| h001 | Keep save mutation allowlisted and validate numeric IDs before rendering them into HTML. | Unknown save fields must survive, and unvalidated values must not cross into `innerHTML`. | Brotato save editor run, 2026-10-02 | 2026-10-02 | active |
| h002 | Treat generated WASM/site output as disposable and build it in CI rather than committing it. | Prevents stale or read-only Nix-store-derived artifacts from entering source control. | Brotato save editor run, 2026-10-02 | 2026-10-02 | active |
