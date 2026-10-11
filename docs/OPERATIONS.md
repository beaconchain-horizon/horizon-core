# Horizon v1.4 — Operations Manual

**Audience**: Sarcheshmeh IT · Judiciary IT / Khorasan
**Scope**: SLA, backup, disaster recovery, training, support

---

## 1. Service Level Agreement (SLA)

### Availability

| Component | Target uptime | Notes |
|---|---|---|
| Backend API | 99.5% / month | ~3.6h downtime / month allowed |
| Switch gateway | 99.5% / month | |
| Ledger storage | 99.9% / month | Postgres or SQLite |

### Response times

| Severity | Definition | Response | Resolution |
|---|---|---|---|
| P1 — Critical | Ledger stopped, licenses rejected | 1h | 4h |
| P2 — Major | Performance degraded >50% | 4h | 24h |
| P3 — Minor | Single feature impaired | 24h | 7d |
| P4 — Cosmetic | UI/typo | 7d | next release |

### Support channels

- Email: support@horizon-core (response < 4h business hours)
- Phone: emergency line for P1 only
- On-site: for P1 within 24h if remote fails

---

## 2. Backup strategy

### Online mode (PostgreSQL)

```bash
# Daily full backup
pg_dump -Fc horizon_db > /backup/horizon-$(date +%Y%m%d).dump

# Continuous WAL archiving for point-in-time recovery
archive_command = 'cp %p /backup/wal/%f'
```

- **RPO** (Recovery Point Objective): 5 minutes
- **RTO** (Recovery Time Objective): 1 hour

### Offline mode (SQLite)

```bash
# Daily snapshot (safe while running)
sqlite3 /data/horizon.db ".backup '/backup/horizon-$(date +%Y%m%d).db'"
```

- **RPO**: 24 hours (or 1 hour with hourly cron)
- **RTO**: 15 minutes (copy file back)

### Both modes

- **Merkle root backup**: publish the daily Merkle root to a separate
  append-only file `/backup/merkle-roots.log`. This proves that data
  existed on a given day even if the DB is lost.
- **USB export**: monthly, export the ledger as a signed bundle for
  off-site storage.

---

## 3. Disaster recovery

### Scenario 1 — Backend crash

1. Systemd restarts it: `systemctl restart horizon-api`
2. Ledger resumes from the last block
3. **Zero data loss** (WAL persistence)

### Scenario 2 — Disk failure

1. Replace disk
2. Restore latest snapshot
3. Ledger resumes from snapshot
4. **Data loss = up to 5 minutes** (from WAL)
5. Replay from CBS/SCADA if needed (idempotent)

### Scenario 3 — Database corruption

1. Stop services
2. Restore latest backup
3. Verify Merkle root matches the published one
4. **Data loss = up to 24 hours**
5. Replay from source if available

### Scenario 4 — USB key lost

1. Rotate the key: generate new secp256k1 key
2. Publish new public key to Switch
3. Sign an invalidation record with the new key
4. **Prior signed blocks remain valid** (they were signed with the old key)
5. New blocks signed with the new key

### Scenario 5 — Complete site loss

1. Restore from off-site USB export
2. Rebuild from CBS/SCADA if the source has the full history
3. **Recovery from cold backup < 4 hours**

---

## 4. Training

### Operator training (2 days)

Day 1:
- Architecture overview (L1-L2 bank / L2.5 industry)
- Starting/stopping services
- Reading logs
- Health checks

Day 2:
- License issuance workflow
- USB key handling
- Backup procedures
- Incident response basics

### Administrator training (3 additional days)

- Key rotation
- Merkle proof generation for auditors
- Disaster recovery drills
- Configuration tuning
- Security hardening

### Documentation provided

- This operations manual (Persian + English)
- Video walkthroughs (10 x 15 min)
- Runbooks for common incidents

---

## 5. Reporting to regulator

### Daily report (JSON + PDF)

```json
{
  "date": "2026-10-11",
  "transactions_anchored": 128472,
  "merkle_root": "0xabc123...",
  "block_signature": "0xdef456...",
  "authority_public_key": "0x...",
  "chain_height": 9843
}
```

### Monthly digest

- Aggregated stats
- Merkle root history
- Key rotation events (if any)
- Incident log

### On-demand proof bundle

For any single transaction:
- Canonical payload
- Merkle inclusion proof (path)
- Block signature
- Public key
- Verification command

The regulator can verify this **offline** with `cmd/licenseverify`.

---

## 6. Monitoring

### Health endpoints

| Endpoint | Purpose |
|---|---|
| `GET /health` | Liveness |
| `GET /api/v1/health` | Full health (DB + ledger) |
| `GET /stats` | Metrics |

### Alerts

- Ledger stopped producing blocks > 5 min → P1
- Signature verification failure rate > 0 → P1
- Disk usage > 85% → P2
- DB replication lag > 60s → P2

### Log retention

- Application logs: 90 days
- Audit logs (Merkle roots): forever
- Access logs: 1 year

---

## 7. Upgrade procedure

1. Download signed release bundle
2. Verify signature with the current public key
3. Stop services
4. Backup DB
5. Replace binaries
6. Run migrations
7. Start services
8. Verify health + smoke test
9. Rollback plan: keep previous binaries for 7 days

---

## 8. Costs (indicative)

| Item | One-time | Recurring |
|---|---|---|
| Installation | included | — |
| Training | included | — |
| Annual license | — | per contract |
| Support (business hours) | — | included |
| Support (24/7) | — | optional |
| On-site visits | travel | per visit |

---

**© 2026 Horizon — Mahdi Amoli Moghaddam**
