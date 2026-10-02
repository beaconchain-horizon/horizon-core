import { NextResponse } from 'next/server';

export async function DELETE() {
  const response = NextResponse.json({ ok: true });

  response.cookies.set('angel_session', '', {
    httpOnly: true,
    secure: process.env.NODE_ENV === 'production',
    sameSite: 'lax',
    path: '/',
    expires: new Date(0),
  });

  return response;
}
