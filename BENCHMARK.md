# Horizon Core — Benchmark Results

Date: 2026-10-02
Environment: Windows 10, 8-core, GOMAXPROCS=8
Test: POST /api/v1/tx/batch

## Peak (preliminary)

- Ledger only:  241,545 TPS
- With chain:   124,533 TPS
- Batch size:   100,000 TX per request

## IMPORTANT LIMITATION

The test did NOT flush the full 100,000 TX.

- Submitted: 100,000
- Flushed:   1,840 (in 8s window)
- Auto-flush rate limited by async batcher contention

The 124,533 TPS figure is NOT sustained end-to-end throughput
for a complete 100,000 TX workload.

What it DOES show:
- Ledger processes 100K TX in < 500ms
- Block creation + signing works
- No data loss in flushed window
- System stable under load

## Historical

- 2026-09  Industrial reading:  32,849 TPS
- 2026-10  Banking batch:       124,533 TPS (preliminary)
- 2026-10  Banking single:      ~3 TPS (curl-bound)

## Known Issues

1. Async batcher contention - needs batcher v2 design
2. Single-TX ~300ms - needs adaptive batching

## Reproduce

The throughput values above are historical and preliminary; they are not a claim of sustained end-to-end throughput. The local prebuilt `./bin/switch.exe` binary is not part of this source repository.

Use Go 1.24 for source validation:

```bash
go test -count=1 ./...
```

A new throughput figure must be collected against a fully configured switch, registered sensor, and compatible test key. Record submitted/flushed counts and errors alongside TPS.
