import { NextResponse } from 'next/server';

export async function POST(req: Request) {
  const endpoint = process.env.HORIZON_ORDER_API || '';
  const apiKey = process.env.HORIZON_ORDER_API_KEY || '';
  if (!endpoint || apiKey.length < 32) {
    return NextResponse.json({
      ok: false,
      error: 'order_service_not_configured',
      message: 'Order processing is not configured; no demo order was created.',
    }, { status: 503 });
  }

  const body = await req.json().catch(() => null);
  if (!body || typeof body !== 'object' || Array.isArray(body)) {
    return NextResponse.json({ ok: false, error: 'invalid_order' }, { status: 400 });
  }

  try {
    const upstream = await fetch(endpoint, {
      method: 'POST',
      headers: {
        'content-type': 'application/json',
        'X-API-Key': apiKey,
      },
      body: JSON.stringify(body),
      cache: 'no-store',
      signal: AbortSignal.timeout(10000),
    });
    const data = await upstream.json().catch(() => ({}));
    return NextResponse.json(data, { status: upstream.status });
  } catch {
    return NextResponse.json({
      ok: false,
      error: 'order_service_unavailable',
      message: 'The configured Horizon order service is unavailable.',
    }, { status: 502 });
  }
}
