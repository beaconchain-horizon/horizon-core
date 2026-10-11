# Horizon v1.4 — Deployment Requirements & Integration

**Audience**: Sarcheshmeh · Judiciary / Khorasan
**Scope**: installation, hardware, network, CBS/SCADA integration
**Mode**: offline (SQLite) · online (PostgreSQL)

---

## 1. Hardware & OS requirements

### Minimal production

| Resource | Minimum | Recommended |
|---|---|---|
| CPU | 4 core x86_64 | 8 core |
| RAM | 8 GB | 16 GB |
| Disk | 100 GB SSD | 500 GB NVMe |
| Network | internal LAN only | internal LAN + backup link |
| OS | Ubuntu 22.04 LTS or Windows Server 2019 | Ubuntu 24.04 LTS |

### For air-gapped sites

- **No internet required.** All dependencies bundled in the installer.
- **No outbound calls.** Switch and Backend refuse external connections by default.
- **USB transfer** for updates: signed bundle, verified with the same public key.

---

## 2. Two operating modes

### Mode A — Online (PostgreSQL)

```
DATABASE_URL=postgresql://horizon:password@db-server:5432/horizon_db
```

- Persistent storage on a dedicated DB server
- Replication supported (primary + standby)
- Recommended for central offices and data centers

### Mode B — Offline (SQLite)

```
DATABASE_URL=                          # left empty
HORIZON_DB_PATH=/data/horizon.db       # optional, default ./horizon-dev.db
```

- Single-file database, WAL mode
- No external DB process
- Recommended for remote sites, air-gap, single-node installs
- Daily snapshot: `sqlite3 horizon.db ".backup 'snapshot-YYYYMMDD.db'"`

Both modes are **feature-identical**. The same license, the same Switch,
the same Merkle proofs.

---

## 3. CBS integration (Core Banking)

### Connection method

| Scenario | Method |
|---|---|
| CBS has REST API | Backend polls `GET /api/v1/transactions?since=<ts>` every 30s |
| CBS has message queue | Kafka / RabbitMQ consumer, topic `tx.settled` |
| CBS has only files | CSV/SFTP drop, watcher reads every 60s |
| CBS has none | Manual import via authenticated endpoint |

### Data flow

```
CBS (read-only source)
       │
       │  read every 30s / stream
       ▼
   Horizon Backend
       │
       │  canonicalize + hash
       ▼
   Merkle batcher → block every 30s
       │
       │  sign with Authority (USB)
       ▼
   Immutable ledger (Postgres or SQLite)
```

### Canonical format for a transaction

```
cbs_tx_id|account_id|amount_minor|currency|direction|timestamp|reference
```

- `amount_minor`: integer in minor units (1 IRR = 1 unit; 1 USD = 100 cents)
- `direction`: `DR` or `CR`
- `timestamp`: RFC3339 UTC, no milliseconds

### Guarantees

- Horizon **never writes to CBS**. Read-only.
- If CBS is unavailable, Horizon buffers locally and resumes.
- No change to CBS code required.

---

## 4. SCADA / Historian integration (industrial)

### Connection method

| Source | Protocol | Adapter |
|---|---|---|
| SCADA (OPC-UA) | OPC-UA | built-in client |
| SCADA (Modbus TCP) | Modbus | built-in client |
| Historian (SQL) | JDBC/ODBC | scheduled query |
| Custom | REST | POST to `/api/v1/industrial/reading` |

### Anchoring cycle

- Every 30 seconds (configurable)
- Batch up to 50,000 readings per cycle
- Each reading produces a Merkle leaf
- Each batch produces a signed Merkle root

### Guarantees

- No write-back to SCADA/PLC
- Works inside the same OT network segment (L2.5)
- Or across a software diode to L3/L4

---

## 5. Network architecture

### Online deployment

```
┌──────────────┐     ┌──────────────┐     ┌──────────────┐
│  CBS / SCADA │────▶│   Horizon    │────▶│  PostgreSQL  │
│  (read-only) │     │   Backend    │     │  (primary)   │
└──────────────┘     └──────┬───────┘     └──────┬───────┘
                            │                    │
                            │  /license/check    │  replication
                            ▼                    ▼
                     ┌──────────────┐     ┌──────────────┐
                     │   Switch     │     │  PostgreSQL  │
                     │  (policies)  │     │  (standby)   │
                     └──────────────┘     └──────────────┘
```

### Air-gapped deployment

```
┌──────────────┐     ┌──────────────┐
│  SCADA (L2)  │────▶│   Horizon    │
│              │     │  (L2.5)      │
└──────────────┘     └──────┬───────┘
                            │  SQLite (local)
                            ▼
                     ┌──────────────┐
                     │  /data/h.db  │
                     └──────────────┘
                     ↑
        USB export for regulator
```

---

## 6. Installation steps

```bash
# 1. Copy the repo onto the target server
scp -r horizon-core/ user@server:/opt/horizon-core

# 2. Set environment
cd /opt/horizon-core
cp .env.example .env
# edit .env: ADMIN_TOKEN, DATABASE_URL (or leave empty for SQLite)

# 3. Build binaries
go build -o bin/horizon-api     ./cmd/api
go build -o bin/horizon-switch  ./cmd/switch

# 4. Run migrations (Backend does this on startup)
./bin/horizon-api &
./bin/horizon-switch &

# 5. Verify
curl http://localhost:8080/api/v1/health
curl http://localhost:8081/health
```

Installation time: **~1 hour** on a prepared server.

---

## 7. What changes in the CBS / SCADA

**Nothing.** Zero code changes. Zero schema changes. Zero downtime.

Horizon attaches as a read-only observer. If you disconnect Horizon,
the CBS continues exactly as before.

---

## 8. Sizing guide

| Environment | Records/day | Recommended mode | Hardware |
|---|---|---|---|
| Single site (10 sensors) | 100K | SQLite | 4 core / 8 GB |
| Oil refinery | 5M | PostgreSQL | 8 core / 16 GB |
| Central bank settlement | 20M | PostgreSQL + replication | 16 core / 32 GB |
| Air-gapped critical | 1M | SQLite + USB export | 8 core / 16 GB |

---

**© 2026 Horizon — Mahdi Amoli Moghaddam**
