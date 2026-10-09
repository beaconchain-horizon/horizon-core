# Repository scripts

- `install.sh` — Linux/macOS installation helper. Run it from a Git checkout and configure secrets in the generated `.env` before using the services.
- The Go benchmark tool lives at `cmd/tpsbench`; it requires a running switch, a registered sensor, and a matching private test key. Do not use historical TPS values as a replacement for an end-to-end test.

The old root-level Next.js repair, auto-commit/push, and legacy deployment scripts have been removed.
