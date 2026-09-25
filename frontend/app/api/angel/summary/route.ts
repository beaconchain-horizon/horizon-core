import {NextResponse} from 'next/server';
import {config} from '../../../../lib/config';
import {agents} from '../../../../lib/demo-data';
import {cookies} from 'next/headers';
import {validSession} from '../../../../lib/angel-session';
async function fetchJSON(url:string){const r=await fetch(url,{cache:'no-store',signal:AbortSignal.timeout(3500)});return {status:r.status,data:await r.json().catch(()=>({}))}}
export async function GET(){
  const c=await cookies();
  if(!validSession(c.get('hz_angel_session')?.value||null)) return NextResponse.json({error:'unauthorized'},{status:401});
  if(config.switchUrl){try{const base=config.switchUrl.replace(/\/$/,'');const [health,stats,benchmark]=await Promise.all([fetchJSON(`${base}/health`),fetchJSON(`${base}/stats`),fetchJSON(`${base}/benchmark`)]);return NextResponse.json({mode:'live',health,stats,benchmark,agents:agents.length});}catch{}}
  return NextResponse.json({mode:'demo',health:{status:200,data:{status:'online'}},stats:{status:200,data:{chainLength:4821}},benchmark:{status:200,data:{tps:4140}},agents:agents.length});
}
