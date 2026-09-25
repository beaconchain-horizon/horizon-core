import {NextResponse} from 'next/server';
import {cookies} from 'next/headers';
import {validSession} from '../../../../lib/angel-session';
import {config} from '../../../../lib/config';
export async function POST(req:Request){
  const c=await cookies();
  if(!validSession(c.get('hz_angel_session')?.value||null)) return NextResponse.json({error:'unauthorized'},{status:401});
  const body=await req.json().catch(()=>({}));
  if(config.issueApi){try{const r=await fetch(config.issueApi,{method:'POST',headers:{'content-type':'application/json'},body:JSON.stringify(body),cache:'no-store'});const d=await r.json().catch(()=>({}));return NextResponse.json(d,{status:r.status})}catch{}}
  return NextResponse.json({status:'issued_demo',license_id:`LIC-DEMO-${Date.now()}`,binding:{customer_id:body.customer_id,branch_id:body.branch_id,industry_id:body.industry_id,hardware_id:body.hardware_id},signature:'BACKEND_SIGNATURE',public_key:'BACKEND_PUBLIC_KEY',seed:'BACKEND_SEED',merkle_root:'BACKEND_MERKLE_ROOT'});
}
