import { cookies } from 'next/headers';
import { NextResponse } from 'next/server';
import { validSession } from '../../../../lib/angel-session';

type IssueRequest = {
  customer_id?: string;
  branch_id?: string;
  industry_id?: string;
  product_id?: string;
  hardware_id?: string;
  order_id?: string;
  volume?: number;
  duration?: number;
};

function clean(value: unknown, max = 200): string {
  return typeof value === 'string' ? value.trim().slice(0, max) : '';
}

export async function POST(request: Request) {
  const cookieStore = await cookies();
  if (!validSession(cookieStore.get('hz_angel_session')?.value || null)) {
    return NextResponse.json({ ok: false, error: 'owner_auth_required' }, { status: 401 });
  }

  let body: IssueRequest;
  try {
    body = (await request.json()) as IssueRequest;
  } catch {
    return NextResponse.json({ ok: false, error: 'invalid_json' }, { status: 400 });
  }

  const customerID = clean(body.customer_id);
  const branchID = clean(body.branch_id);
  const productID = clean(body.product_id) || clean(body.industry_id);
  const hardwareID = clean(body.hardware_id, 512);
  const volume = Number(body.volume);
  const duration = Number(body.duration);

  if (!customerID || !branchID || !productID || !hardwareID ||
      !Number.isSafeInteger(volume) || volume <= 0 ||
      !Number.isSafeInteger(duration) || duration <= 0) {
    return NextResponse.json({
      ok: false,
      error: 'invalid_license_request',
      message: 'customer_id, branch_id, product_id, hardware_id, positive volume, and positive duration are required',
    }, { status: 400 });
  }

  // These values exist only in the server runtime, never in a browser bundle.
  const issueURL = process.env.HORIZON_ISSUE_API || '';
  const apiKey = process.env.HORIZON_API_KEY || '';
  if (!issueURL || apiKey.length < 32) {
    return NextResponse.json({
      ok: false,
      error: 'authority_not_configured',
      message: 'Secure license issuance is not configured on the server.',
    }, { status: 503 });
  }

  try {
    const upstream = await fetch(issueURL, {
      method: 'POST',
      headers: {
        'content-type': 'application/json',
        'X-API-Key': apiKey,
      },
      body: JSON.stringify({
        product_id: productID,
        user_id: customerID + ':' + branchID,
        volume,
        duration,
        hardware_id: hardwareID,
      }),
      cache: 'no-store',
      signal: AbortSignal.timeout(15000),
    });
    const data = await upstream.json().catch(() => ({})) as {
      license?: { statement?: { license_id?: string }; signature?: string };
      error?: string;
    };
    if (!upstream.ok || !data.license?.statement?.license_id || !data.license.signature) {
      return NextResponse.json({
        ok: false,
        error: 'authority_rejected_request',
        message: data.error || 'Horizon Authority rejected the license request.',
      }, { status: upstream.ok ? 502 : upstream.status });
    }

    return NextResponse.json({
      ok: true,
      source: 'horizon-authority-v1.4',
      license_id: data.license.statement.license_id,
      customer_id: customerID,
      branch_id: branchID,
      product_id: productID,
      hardware_id: hardwareID,
      order_id: clean(body.order_id),
      license: data.license,
    }, { status: 201 });
  } catch {
    return NextResponse.json({
      ok: false,
      error: 'authority_unavailable',
      message: 'The secure Horizon Authority endpoint is unavailable.',
    }, { status: 502 });
  }
}
