# Horizon Store + Horizon Angel SOC (Next.js)

A clean Next.js frontend for the Horizon ecosystem with two clearly separated surfaces:

- Public Store: products, monitoring product, demo, ordering, checkout placeholder, customer portal.
- Private Horizon Angel: owner-only operations console for agents, customers, orders, licenses and monitoring.

## Key workflow

1. Customer chooses a product and quantity.
2. Customer submits an order.
3. Order can later be connected to PayPal/card provider.
4. Owner sees pending orders in Horizon Angel.
5. Owner binds customer/branch/industry/hardware identity and requests license issuance from the secure backend.
6. Private key remains server-side/owner-side and is never sent to the public browser.
7. Customer receives the issued `.lic` payload in their private customer portal.

## Local run

Requires Node.js 20.9+ (Next.js requirement).

```bash
npm install
cp .env.example .env.local
npm run dev
```

Then open http://localhost:3000

## Production integration

Set `HORIZON_API_URL`, `HORIZON_SWITCH_URL`, `HORIZON_ORDER_API`, and `HORIZON_ISSUE_API` to your Liara endpoints. The UI contains demo fallback data when integration endpoints are not configured so the frontend remains reviewable locally.

Do not place the Horizon private key in `.env` files shipped to browsers, localStorage, or public Next.js variables.

## Source references

The content shown on the documents page is based on the Horizon Core repository and the technical certificate at the referenced commit.
