import { NextResponse } from 'next/server';
import { agents } from '../../../../lib/demo-data';

export async function GET() {
  const online = agents.filter((a) => a.status === 'online').length;
  const warning = agents.filter((a) => a.status === 'warning').length;
  const offline = agents.filter((a) => a.status === 'offline').length;

  return NextResponse.json({
    agents: agents.length,
    online,
    warning,
    offline,
    source: 'demo',
    timestamp: new Date().toISOString(),
  });
}
