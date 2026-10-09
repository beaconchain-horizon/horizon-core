import { createHmac, randomBytes, timingSafeEqual } from 'node:crypto';

const SESSION_TTL_SECONDS = 60 * 60;
const secret = () => process.env.ANGEL_SESSION_SECRET || '';

export function sessionToken(): string {
  const key = secret();
  if (key.length < 32) {
    throw new Error('ANGEL_SESSION_SECRET must be configured with at least 32 characters');
  }

  const issuedAt = Math.floor(Date.now() / 1000);
  const expiresAt = issuedAt + SESSION_TTL_SECONDS;
  const nonce = randomBytes(16).toString('hex');
  const payload = `${issuedAt}.${expiresAt}.${nonce}`;
  const signature = createHmac('sha256', key).update(payload).digest('hex');
  return `${payload}.${signature}`;
}

export function validSession(value: string | null): boolean {
  const key = secret();
  if (!value || key.length < 32 || value.length > 256) return false;

  const parts = value.split('.');
  if (parts.length !== 4) return false;
  const [issuedText, expiresText, nonce, suppliedSignature] = parts;
  if (!/^\d+$/.test(issuedText) || !/^\d+$/.test(expiresText) ||
      !/^[a-f0-9]{32}$/.test(nonce) || !/^[a-f0-9]{64}$/.test(suppliedSignature)) {
    return false;
  }

  const issuedAt = Number(issuedText);
  const expiresAt = Number(expiresText);
  const now = Math.floor(Date.now() / 1000);
  if (!Number.isSafeInteger(issuedAt) || !Number.isSafeInteger(expiresAt) ||
      issuedAt > now + 60 || expiresAt <= now ||
      expiresAt <= issuedAt || expiresAt - issuedAt > SESSION_TTL_SECONDS) {
    return false;
  }

  const payload = `${issuedAt}.${expiresAt}.${nonce}`;
  const expected = Buffer.from(createHmac('sha256', key).update(payload).digest('hex'), 'utf8');
  const supplied = Buffer.from(suppliedSignature, 'utf8');
  return expected.length === supplied.length && timingSafeEqual(expected, supplied);
}
