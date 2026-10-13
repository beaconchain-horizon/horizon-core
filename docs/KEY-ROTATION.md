# Key Rotation

## When to rotate

- Scheduled: every 12 months
- On compromise suspicion: immediately
- On team departure with key access: within 24h
- On USB key replacement: before decommission

## Procedure

1. Generate new key air-gapped: `go run ./cmd/horizon-keygen > new-key.pem`
2. Transfer via USB only.
3. Run: `scripts/rotate-signing-key.sh new-key.pem`
4. Apply printed env changes.
5. Restart backend.
6. Append to reports/key-rotations.log.

## Overlap window

Licenses signed with old key remain verifiable only while old public key is retained. Switch must accept both keys during overlap. Recommended: 90 days.

## What we do NOT do

- Never delete old keys. Move to reports/quarantine-YYYYMMDD/.
- Never rotate automatically. Human decision with paper trail.
