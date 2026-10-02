import { NextResponse } from 'next/server';
import { angelAgents } from '../../../../lib/angel-demo-data';

export async function GET() {
  const online = angelAgents.filter(
    (a) => a.status === 'online'
  ).length;

  const warning = angelAgents.filter(
    (a) => a.status === 'warning'
  ).length;

  const offline = angelAgents.filter(
    (a) => a.status === 'offline'
  ).length;

  return NextResponse.json({
    agents: angelAgents.length,
    online,
    warning,
    offline,
    source: 'demo',
    timestamp: new Date().toISOString(),
  });
}
