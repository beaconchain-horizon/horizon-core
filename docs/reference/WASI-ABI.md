# WASI Application ABI

Technical reference — WebAssembly System Interface Application ABI

## Use Cases in Horizon

- **Air-Gap:** Run one binary on any OS without recompilation
- **License Panel:** Use Go code in browser (WASM)
- **Deployment:** Banks on any infrastructure (Linux / macOS / FreeBSD / WASI runtime)

## Sources

- https://github.com/WebAssembly/WASI/issues/13
- https://github.com/WebAssembly/WASI/issues/19
- https://github.com/WebAssembly/WASI/issues/24

## ABI Summary

### Two Module Types

**Command:**
- Exports `_start` function (no args, no return)
- Default export
- Called at most once

**Reactor:**
- Exports `_initialize` (optional)
- After call, instance stays live
- Exports accessible

### Required Exports (all modules)
- `memory` — linear memory
- `__indirect_function_table` — function table

### File Descriptors
- `0` = stdin
- `1` = stdout
- `2` = stderr
- preopens: via `fd_prestat_get` and `fd_prestat_dir_name`

### Forbidden Exports
- `__heap_base`
- `__data_end`

## Status
- Current ABI: **Unstable**
- Stable ABI: **Under discussion**
