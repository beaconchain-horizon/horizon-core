import { cookies } from 'next/headers';
import { NextResponse } from 'next/server';
import { validSession } from '../../../../lib/angel-session';

export async function GET() {
  const cookieStore = await cookies();
  if (!validSession(cookieStore.get('hz_angel_session')?.value || null)) {
    return NextResponse.json({ ok: false, error: 'owner_auth_required' }, { status: 401 });
  }

  let switchURL = (process.env.HORIZON_SWITCH_URL || '').trim();
  while (switchURL.endsWith('/')) switchURL = switchURL.slice(0, -1);
  if (!switchURL) {
    return NextResponse.json({
      ok: false,
      source: 'unconfigured',
      status: 'unconfigured',
      message: 'HORIZON_SWITCH_URL is not configured on the server.',
    }, { status: 503 });
  }

  try {
    const [healthResponse, statsResponse] = await Promise.all([
      fetch(switchURL + '/health', { cache: 'no-store', signal: AbortSignal.timeout(5000) }),
      fetch(switchURL + '/stats', { cache: 'no-store', signal: AbortSignal.timeout(5000) }).catch(() => null),
    ]);
    const health = await healthResponse.json().catch(() => ({})) as Record<string, unknown>;
    const stats = statsResponse?.ok
      ? await statsResponse.json().catch(() => ({})) as Record<string, unknown>
      : null;

    if (!healthResponse.ok || health.status !== 'online') {
      return NextResponse.json({
        ok: false,
        source: 'switch',
        status: 'unavailable',
        health,
      }, { status: 502 });
    }

    return NextResponse.json({
      ok: true,
      source: 'horizon-switch',
      status: 'online',
      version: typeof health.version === 'string' ? health.version : 'unknown',
      service: typeof health.service === 'string' ? health.service : 'horizon-switch',
      timestamp: new Date().toISOString(),
      stats,
    });
  } catch {
    return NextResponse.json({
      ok: false,
      source: 'switch',
      status: 'unavailable',
      message: 'Horizon Switch could not be reached.',
    }, { status: 502 });
  }
}
