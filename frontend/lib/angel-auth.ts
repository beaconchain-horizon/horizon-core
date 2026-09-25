import {cookies} from 'next/headers';
import {redirect} from 'next/navigation';
import {validSession} from './angel-session';
export async function requireAngel(){const c=await cookies();if(!validSession(c.get('hz_angel_session')?.value||null))redirect('/angel/login');}
