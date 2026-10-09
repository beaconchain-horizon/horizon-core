# Horizon Core v1.4 — Recorded Legacy Regression Evidence

Date: 2026-10-09 (local integration run supplied before v1.4 protocol migration).

This is a sanitized record of the observed local integration test. Credentials, signature bytes, license IDs, seed material, and public-key contents are deliberately not copied here.

## Observed service routing

| Local endpoint | Observed response |
|---|---|
| `GET /api/v1/health` on port `8080` | Backend reported `online` |
| `GET /health` on port `8081` | Switch reported `online`, version `1.4` |

## Observed legacy license contract

The running deployment generated a license with the fields `id`, `volume`, `root`, `seed`, `signature`, `public_key`, `signed_data`, and `expires_at`. The legacy `signed_data` was a pipe-delimited string with seven fields. This is evidence for the old interface only, not the new canonical v1.4 envelope.

Observed checks:

- A generated license sent to the Switch `/api/v1/license/check` returned `valid: true` and `verified: true`.
- A signature altered by changing one hex digit returned `valid: false`, `verified: false`, and `signature verification failed`.
- The signed data altered from volume `1000` to `999` returned `valid: false`, `verified: false`, and `signature verification failed`.

## Migration rule

Keep this as baseline evidence. The unified protocol in this branch signs one structured `statement` envelope, binds a required Hardware ID, validates an immutable-field Merkle root, and expects the Switch's Authority public key to be provisioned independently. Do not silently reinterpret the legacy pipe-delimited message as a valid new-format license. Any compatibility path for legacy licenses must be explicit, labelled legacy, and not used for operations requiring the new HWID-bound protocol.

The pre-migration test is not being represented as proof that the new protocol has been exercised against live Backend and Switch containers. Go CI tests and the live end-to-end integration run are separate evidence.
