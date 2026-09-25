import {createHmac,timingSafeEqual} from 'node:crypto';

const secret=()=>process.env.ANGEL_SESSION_SECRET||'dev-only-secret-change-me';
export const sessionToken=()=>createHmac('sha256',secret()).update('horizon-angel-session').digest('hex');
export function validSession(value:string|null){if(!value)return false;const a=Buffer.from(value);const b=Buffer.from(sessionToken());return a.length===b.length&&timingSafeEqual(a,b)}
