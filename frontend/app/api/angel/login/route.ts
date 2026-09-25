import {NextResponse} from 'next/server';
import {sessionToken} from '../../../../lib/angel-session';

export async function POST(req:Request){
  const b=await req.json().catch(()=>({}));
  const configured=process.env.ANGEL_ADMIN_PASSWORD||'change-me';
  if(String(b.password||'')!==configured) return NextResponse.json({error:'invalid credentials'},{status:401});
  const res=NextResponse.json({ok:true});
  res.cookies.set('hz_angel_session',sessionToken(),{httpOnly:true,secure:process.env.NODE_ENV==='production',sameSite:'lax',path:'/',maxAge:60*60*8});
  return res;
}
export async function DELETE(){
  const res=NextResponse.json({ok:true});
  res.cookies.set('hz_angel_session','',{httpOnly:true,secure:process.env.NODE_ENV==='production',sameSite:'lax',path:'/',maxAge:0});
  return res;
}
