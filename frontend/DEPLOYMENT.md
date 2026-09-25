# Local → Liara/domain checklist

1. Copy `.env.example` to `.env.local`.
2. Set `NEXT_PUBLIC_SITE_URL` to the purchased domain.
3. Set `HORIZON_API_URL` and `HORIZON_SWITCH_URL` to your real Liara API/switch hosts.
4. Set `HORIZON_ORDER_API` once the order backend endpoint is ready.
5. Set `HORIZON_ISSUE_API` to the secure license issuance endpoint.
6. Set `ANGEL_ADMIN_PASSWORD` and a long random `ANGEL_SESSION_SECRET`.
7. Connect PayPal/card credentials on the server only. Never put private key material in `NEXT_PUBLIC_*` variables.
8. Run `npm run build && npm start` for production verification.

The UI intentionally works with demo fallback data when the API variables are blank. This makes design review possible before the payment and issuance infrastructure is connected.
