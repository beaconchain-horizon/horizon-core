# Horizon Store / Angel architecture

## Public

`/` — clean storefront landing page
`/products` — product catalog
`/products/monitoring` — industrial monitoring product
`/monitor-demo` — public interactive demo
`/cart` — cart
`/checkout` — order creation; payment provider placeholder
`/customer` — customer portal UI
`/about` — product architecture and workflow
`/documents` — technical documents/certificates from GitHub

## Private

`/angel/login` — owner authentication gate
`/angel` — private owner SOC UI

The private key is deliberately absent from all public frontend flows. The issuance API is server-side and can be connected to the Horizon secure backend through `HORIZON_ISSUE_API`.

## API adapters

`/api/store/orders` — forwards to `HORIZON_ORDER_API` when configured, otherwise returns a demo order.
`/api/angel/summary` — reads Switch health/stats/benchmark when `HORIZON_SWITCH_URL` is configured, otherwise returns demo telemetry.
`/api/angel/issue-license` — forwards license issuance to `HORIZON_ISSUE_API` when configured; otherwise returns clearly-marked demo issuance metadata.
`/api/angel/login` — creates an HttpOnly owner session without exposing private key material.
