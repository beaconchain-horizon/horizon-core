import { timingSafeEqual } from 'node:crypto';
import { NextResponse } from 'next/server';
import { sessionToken } from '../../../../lib/angel-session';

export async function POST(request: Request) {
  const expected = process.env.ANGEL_ADMIN_PASSWORD || '';
  const sessionSecret = process.env.ANGEL_SESSION_SECRET || '';
  if (expected.length < 32 || sessionSecret.length < 32 || expected === 'CHANGE_ME') {
    return NextResponse.json(
      { ok: false, error: 'owner_auth_not_configured' },
      { status: 503 }
    );
  }

  const body = await request.json().catch(() => null) as { password?: unknown } | null;
  const supplied = typeof body?.password === 'string' ? body.password : '';
  const left = Buffer.from(supplied, 'utf8');
  const right = Buffer.from(expected, 'utf8');
  const matches = left.length === right.length && timingSafeEqual(left, right);
  if (!matches) {
    return NextResponse.json({ ok: false, error: 'invalid_credentials' }, { status: 401 });
  }

  const response = NextResponse.json({ ok: true });
  response.cookies.set('hz_angel_session', sessionToken(), {
    httpOnly: true,
    secure: process.env.NODE_ENV === 'production',
    sameSite: 'strict',
    path: '/',
    maxAge: 60 * 60,
  });
  return response;
}

export async function DELETE() {
  // Always clear the cookie, even if expired or malformed.
  const response = NextResponse.json({ ok: true });
  response.cookies.set('hz_angel_session', '', {
    httpOnly: true,
    secure: process.env.NODE_ENV === 'production',
    sameSite: 'strict',
    path: '/',
    expires: new Date(0),
    maxAge: 0,
  });
  return response;
}
