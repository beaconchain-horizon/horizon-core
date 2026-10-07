# Horizon Panels Specification

Version: v1.0
Date: 2026-10-07
Owner: Mahdi Amoli Moghaddam
Purpose: Reference for agents building Horizon panels

---

## Overview

Three independent panels:

| Panel | Location | User | Database |
|-------|----------|------|----------|
| 1 | Customer server (next to switch) | Customer IT admin | SQLite |
| 2 | Liara | Owner + support team | PostgreSQL |
| 3 | Mobile/Web | End user | (from panels 1 and 2) |

### Shared stack

- Next.js 14 (App Router)
- TypeScript
- Tailwind CSS + RTL
- Vazirmatn font
- React Query for state
- WebSocket for live updates
- Light/Dark/System theme
- Recharts for charts

---

## Panel 1 - Customer (Offline Switch)

Location: Customer server, port 8080 (next to switch)
User: Customer IT admin
Status: Offline, no internet

### Page 1: Dashboard

- Total balance from Merkle tree
- Today's transaction count
- License status (days remaining)
- Pending count in catch
- Latest block height
- TPS chart for last 24h

### Page 2: Transactions

- Live list (WebSocket)
- Filters: date, amount, user, merchant
- Search
- Each transaction with: tx_id, merkle_root, signature, block_hash
- Verify button next to each

### Page 3: Merkle Tree

- Graphical tree view
- Fanout selector: 2 or 3
- Click on any leaf shows proof path
- Display merkle_root
- JSON export for audit

### Page 4: License

- Current status
- Expiry date
- Device HWID
- Request renewal from center
- Receive new .lic file
- License history

### Page 5: Sync

- Pending transaction count
- Last sync (time, count)
- Manual sync button
- Sync log
- Sync errors

### Page 6: Accounts

- Customer list
- Balance of each from Merkle
- Search
- Account details

### Page 7: Audit Trail

- All events
- Immutable
- With hash chain
- Search and filter

### Consumed API (from switch)

GET  /api/v1/stats
GET  /api/v1/blocks?limit=&offset=
GET  /api/v1/tx/{id}
GET  /api/v1/proof/{tx_id}?fanout=
POST /api/v1/anchor?fanout=
GET  /api/v1/issuer/pubkey
GET  /api/v1/audit/log
GET  /api/v1/audit/verify/{id}

---

## Panel 2 - Center (Online Backend)

Location: Liara
User: Owner + support team
Status: Online, PostgreSQL

### Page 1: Global Dashboard

- List of all customers
- Each customer: name, HWID, license, last sync, transaction count
- Online/offline status of each switch
- Iran map (optional)
- Global charts

### Page 2: Customers

- CRUD customers
- Info: name, type (bank/industry/judiciary), HWID, contact
- Active licenses
- Sync history

### Page 3: Licenses

- Issue new license (select customer, duration, features)
- Revoke
- Renew
- Download .lic
- History

### Page 4: Sync Received

- Data received from each switch
- With Merkle proof
- Auto verification
- Diff view with last snapshot
- Error log

### Page 5: Mining

- Confirmed blocks
- merkle_root, signature, signer_addr
- Confirmation time
- Pending blocks
- Manual approve

### Page 6: Audit

- Search all audit trails
- Signature verification
- CSV export
- Filter by customer/date

### Page 7: Alerts

- Suspicious transactions (high amount, repeats)
- License expiring (< 30 days)
- Switch offline > 24h
- Sync errors
- Telegram / Email

### Page 8: Reports

- Daily TPS
- Bandwidth usage
- Active customers
- Revenue (license count x price)

### Consumed API (from backend)

POST /api/v1/license/issue
GET  /api/v1/license/verify/{id}
POST /api/v1/license/revoke
POST /api/v1/license/renew
GET  /api/v1/tenants
GET  /api/v1/tenants/{id}
POST /api/v1/sync/receive
GET  /api/v1/sync/received
GET  /api/v1/alerts
GET  /api/v1/reports

---

## Panel 3 - End User

Location: Mobile (PWA) + Web
User: End user (BNPL buyer, bank customer)

### Page 1: Wallet

- Current balance
- Remaining credit
- Current debt
- Income/expense chart

### Page 2: BNPL Purchase

- Select merchant
- Select product
- Select plan (installment count)
- Show installment amount
- Confirm
- Receive Merkle receipt

### Page 3: Transactions

- List
- Details of each transaction
- Link to Merkle proof
- Link to block

### Page 4: Installments

- Installment list
- Due date
- Paid
- Remaining
- Early payment

### Page 5: Profile

- User info
- KYC
- Credit score
- Mobile number
- Settings

### Consumed API

GET  /api/v1/wallet/balance
POST /api/v1/tx/create
GET  /api/v1/user/{user_id}/licenses
GET  /api/v1/proof/{tx_id}
GET  /api/v1/installments
GET  /api/v1/profile

---

## Development Priorities

| Priority | Panel | Reason |
|----------|-------|--------|
| 1 | Panel 1 (Customer) | Bank Tejarat demo on Thursday |
| 2 | Panel 2 (Center) | Owner control |
| 3 | Panel 3 (End user) | After approval |

---

## Security Notes

- All APIs with X-Admin-Token or JWT
- HTTPS mandatory
- Rate limit on all endpoints
- CSRF protection
- XSS prevention (React built-in)
- No API key in frontend
- No private key on browser
- Single-use session token
- Log every request
- CSP header

## Language

- UI: Persian, RTL
- Numbers: Persian digits
- Dates: Jalali (Shamsi)
- Code: English
- Comments: Persian
- Docs: Persian

## Owner

Mahdi Amoli Moghaddam - Owner of Horizon Core and Horizon Angel

--- Horizon Angel - under command of Mahdi Amoli Moghaddam
