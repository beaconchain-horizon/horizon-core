# Horizon Core v1.4 — Security Scan Baseline

Status recorded during the October 9, 2026 v1.4 integration work.

## Toolchain constraint

This branch is intentionally kept on Go 1.24.x as requested. The CI runner used Go 1.24.13. That toolchain constraint has a security consequence: the Go standard-library advisories reported by the scanner included fixes available only in Go 1.25.x. A module-level update cannot patch standard-library code in an already-built Go 1.24 binary.

## Findings previously observed

The repository's previous `govulncheck` log reported 33 vulnerabilities across 3 modules and the Go standard library. Examples in that captured output included:
- `github.com/jackc/pgx/v5@v5.3.1`, with the scanner reporting newer fixed versions such as v5.9.2.
- `golang.org/x/net@v0.33.0`, with the scanner reporting fixes in later versions.
- `github.com/gin-contrib/cors@v1.4.0`, with a later fixed release reported.
- Several standard-library findings in Go 1.24.13 with fixes listed in Go 1.25.x.

This is a record of observed scanner output, not a claim that every finding is exploitable in every deployment. Applicability and reachable paths require review, but no clean vulnerability scan is claimed.

## Release gate

The v1.4 protocol migration does not remove these pre-existing dependency/toolchain findings. The security workflow on this branch therefore reports scanner failure without converting it to success using `|| true`, `-no-fail`, or a silently ignored upload. With Go 1.24 pinned, the expected report may remain red until vulnerable dependency paths are eliminated or the toolchain/backported fixes are explicitly approved. Do not deploy this build to a sensitive production environment on the strength of a passing unit-test suite alone.
