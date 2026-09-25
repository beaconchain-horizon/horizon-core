import {NextResponse} from 'next/server';
import {config} from '../../../../lib/config';

export async function POST(req:Request){
  const body=await req.json().catch(()=>({}));
  if(config.orderApi){
    try{
      const r=await fetch(config.orderApi,{method:'POST',headers:{'content-type':'application/json'},body:JSON.stringify(body),cache:'no-store'});
      const data=await r.json().catch(()=>({}));
      return NextResponse.json(data,{status:r.status});
    }catch{}
  }
  const id=`ORD-${Date.now().toString(36).toUpperCase()}`;
  return NextResponse.json({status:'pending_issuance',order_id:id,mode:'demo',received:body},{status:201});
}
