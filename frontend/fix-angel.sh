#!/usr/bin/env bash
set -e

echo "========================================"
echo " Horizon Angel - Full Repair Script"
echo "========================================"

ROOT="$(pwd)"

if [ ! -f package.json ]; then
  echo "ERROR: package.json پیدا نشد."
  echo "این اسکریپت را در ریشه پروژه Next.js اجرا کن."
  exit 1
fi

echo
echo "[1/8] بررسی ساختار پروژه..."

mkdir -p app/angel
mkdir -p app/api/angel/summary
mkdir -p app/api/angel/issue-license
mkdir -p app/api/angel/login
mkdir -p lib

echo "[2/8] بررسی demo-data..."

if [ ! -f lib/demo-data.ts ]; then
cat > lib/demo-data.ts <<'DATAEOF'
export type AgentStatus = 'online' | 'offline' | 'warning';

export const agents = Array.from({ length: 144 }, (_, i) => {
  const n = String(i + 1).padStart(3, '0');

  return {
    id: `agent-${n}`,
    name: `Horizon Agent ${n}`,
    site: `Site ${n}`,
    status:
      i % 11 === 0
        ? 'offline'
        : i % 7 === 0
          ? 'warning'
          : 'online',
  } satisfies {
    id: string;
    name: string;
    site: string;
    status: AgentStatus;
  };
});

export const customers = [
  {
    id: 'customer-demo-01',
    name: 'Demo Customer',
    type: 'Industrial',
    branches: 1,
    status: 'active',
  },
  {
    id: 'customer-demo-02',
    name: 'Demo Bank',
    type: 'Bank',
    branches: 2,
    status: 'active',
  },
];

export const demoOrders = [
  {
    id: 'ORD-DEMO-001',
    customer: 'Demo Customer',
    product: 'Industrial Monitoring',
    qty: 1,
    status: 'pending_issuance',
  },
  {
    id: 'ORD-DEMO-002',
    customer: 'Demo Bank',
    product: 'Bank Settlement',
    qty: 2,
    status: 'pending_issuance',
  },
];
DATAEOF
fi

echo "[3/8] ساخت API summary..."

cat > app/api/angel/summary/route.ts <<'APIEOF'
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
APIEOF

echo "[4/8] ساخت API صدور لایسنس..."

cat > app/api/angel/issue-license/route.ts <<'APIEOF'
import { NextResponse } from 'next/server';
import crypto from 'crypto';

type IssueRequest = {
  customer_id?: string;
  branch_id?: string;
  industry_id?: string;
  hardware_id?: string;
  order_id?: string;
};

export async function POST(request: Request) {
  try {
    const body = (await request.json()) as IssueRequest;

    if (
      !body.customer_id ||
      !body.branch_id ||
      !body.industry_id ||
      !body.hardware_id ||
      !body.order_id
    ) {
      return NextResponse.json(
        {
          error: 'missing_required_fields',
          message: 'تمام فیلدهای صدور لایسنس الزامی هستند.',
        },
        { status: 400 }
      );
    }

    const licenseId =
      'HZ-' +
      crypto.randomBytes(8).toString('hex').toUpperCase();

    /*
     * DEMO ONLY
     *
     * این endpoint فعلاً UI flow را تست می‌کند.
     * private key نباید در browser یا این route آزمایشی قرار گیرد.
     *
     * در اتصال واقعی:
     *   customer_id
     *   branch_id
     *   industry_id
     *   hardware_id
     *   order_id
     *
     * باید به سرویس صدور واقعی Horizon Core منتقل شوند.
     */

    return NextResponse.json({
      ok: true,
      license_id: licenseId,
      customer_id: body.customer_id,
      branch_id: body.branch_id,
      industry_id: body.industry_id,
      hardware_id: body.hardware_id,
      order_id: body.order_id,
      mode: 'demo',
    });
  } catch {
    return NextResponse.json(
      {
        error: 'invalid_request',
        message: 'JSON درخواست معتبر نیست.',
      },
      { status: 400 }
    );
  }
}
APIEOF

echo "[5/8] ساخت login API..."

cat > app/api/angel/login/route.ts <<'APIEOF'
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
APIEOF

echo "[6/8] اصلاح صفحه Angel..."

cat > app/angel/page.tsx <<'PAGEEOF'
'use client';

import { agents, customers, demoOrders } from '../../lib/demo-data';
import { useEffect, useMemo, useState } from 'react';

type Order = (typeof demoOrders)[number];

type LiveSummary = {
  agents: number;
  online: number;
  warning: number;
  offline: number;
  source?: string;
  timestamp?: string;
};

export default function Angel() {
  const [orders, setOrders] = useState<Order[]>(demoOrders);
  const [live, setLive] = useState<LiveSummary | null>(null);
  const [selected, setSelected] = useState<Order | null>(null);
  const [msg, setMsg] = useState('');
  const [busy, setBusy] = useState(false);

  useEffect(() => {
    let mounted = true;

    fetch('/api/angel/summary', {
      cache: 'no-store',
    })
      .then((r) => {
        if (!r.ok) throw new Error('summary request failed');
        return r.json();
      })
      .then((data: LiveSummary) => {
        if (mounted) setLive(data);
      })
      .catch(() => {
        if (mounted) setLive(null);
      });

    return () => {
      mounted = false;
    };
  }, []);

  const stats = useMemo(() => {
    return {
      online: live?.online ?? agents.filter((a) => a.status === 'online').length,
      warning:
        live?.warning ?? agents.filter((a) => a.status === 'warning').length,
      offline:
        live?.offline ?? agents.filter((a) => a.status === 'offline').length,
      total: live?.agents ?? agents.length,
    };
  }, [live]);

  const issue = async () => {
    if (!selected || busy) return;

    setBusy(true);
    setMsg('');

    try {
      const r = await fetch('/api/angel/issue-license', {
        method: 'POST',
        headers: {
          'content-type': 'application/json',
        },
        body: JSON.stringify({
          customer_id: selected.customer,
          branch_id: 'branch-demo-01',
          industry_id: 'industrial-monitoring',
          hardware_id: 'HW-BIND-DEMO-01',
          order_id: selected.id,
        }),
      });

      const d = await r.json();

      if (!r.ok) {
        throw new Error(d?.message || 'صدور لایسنس ناموفق بود.');
      }

      setMsg(`صدور موفق: ${d.license_id}`);

      setOrders((current) =>
        current.map((x) =>
          x.id === selected.id
            ? { ...x, status: 'issued' }
            : x
        )
      );

      setSelected(null);
    } catch (error) {
      setMsg(
        error instanceof Error
          ? error.message
          : 'خطای ناشناخته در صدور لایسنس'
      );
    } finally {
      setBusy(false);
    }
  };

  return (
    <main>
      <section className="page-head">
        <div className="container">
          <div className="split">
            <div>
              <span className="eyebrow">
                <span className="dot" /> PRIVATE OWNER CONSOLE
              </span>

              <h1>Horizon Angel · SOC</h1>

              <p>
                کنترل مرکزی برای Agentها، مشتری‌ها، سفارش‌ها و صدور
                لایسنس. داده‌های عملیاتی باید از API امن Horizon بیایند.
              </p>
            </div>

            <button
              className="btn btn-danger"
              onClick={async () => {
                try {
                  await fetch('/api/angel/login', {
                    method: 'DELETE',
                  });
                } finally {
                  window.location.href = '/angel/login';
                }
              }}
            >
              خروج
            </button>
          </div>
        </div>
      </section>

      <section className="section">
        <div className="container grid-4">
          <div className="card kpi">
            <div className="v">{stats.total}</div>
            <div className="l">Agentها</div>
          </div>

          <div className="card kpi">
            <div className="v ok-text">{stats.online}</div>
            <div className="l">Online</div>
          </div>

          <div className="card kpi">
            <div className="v warn-text">{stats.warning}</div>
            <div className="l">Warning</div>
          </div>

          <div className="card kpi">
            <div className="v danger-text">{stats.offline}</div>
            <div className="l">Offline</div>
          </div>
        </div>
      </section>

      <section className="section">
        <div className="container">
          <div className="section-title">
            <div>
              <h2>Agent Network</h2>
              <p>
                وضعیت Agentها از API خوانده می‌شود؛ در حالت فعلی داده‌ها
                Demo هستند.
              </p>
            </div>
          </div>

          <div className="agent-grid">
            {agents.map((a) => (
              <div className="agent" key={a.id}>
                <b>{a.name}</b>
                <small>{a.site}</small>

                <div
                  className={`state ${
                    a.status === 'offline'
                      ? 'off'
                      : a.status === 'warning'
                        ? 'warn'
                        : ''
                  }`}
                >
                  <i /> {a.status}
                </div>
              </div>
            ))}
          </div>
        </div>
      </section>

      <section className="section">
        <div className="container two-col">
          <div className="card panel">
            <div className="split">
              <div>
                <h2 style={{ margin: '0 0 5px' }}>
                  Pending issuance
                </h2>

                <div className="muted small">
                  سفارش‌های آماده بررسی و صدور
                </div>
              </div>
            </div>

            <div className="spaced">
              {orders.map((o) => (
                <div
                  className="order-card"
                  key={o.id}
                  style={{
                    border: '1px solid var(--line)',
                    borderRadius: 16,
                    marginBottom: 10,
                    padding: 14,
                  }}
                >
                  <div className="split">
                    <div>
                      <b>{o.customer}</b>

                      <div className="muted small">
                        {o.product} · {o.qty} license
                      </div>
                    </div>

                    <span
                      className={`badge ${
                        o.status === 'issued'
                          ? 'ok'
                          : o.status === 'pending_issuance'
                            ? 'warn'
                            : ''
                      }`}
                    >
                      {o.status}
                    </span>
                  </div>

                  <div className="split">
                    <span className="muted small mono">
                      {o.id}
                    </span>

                    {o.status === 'pending_issuance' && (
                      <button
                        className="btn btn-primary"
                        onClick={() => setSelected(o)}
                      >
                        صدور لایسنس
                      </button>
                    )}
                  </div>
                </div>
              ))}
            </div>

            {msg && (
              <div className="notice ok-text">
                {msg}
              </div>
            )}
          </div>

          <div className="card panel">
            <h2 style={{ marginTop: 0 }}>Customers</h2>

            {customers.map((c) => (
              <div className="status-row" key={c.id}>
                <div>
                  <b>{c.name}</b>

                  <div className="muted small">
                    {c.type} · {c.branches} branch
                  </div>
                </div>

                <span
                  className={
                    c.status === 'active'
                      ? 'ok-text'
                      : 'warn-text'
                  }
                >
                  {c.status}
                </span>
              </div>
            ))}

            <div className="hr" />

            <div className="notice">
              hardware_id باید در binding واقعی لایسنس قرار بگیرد و
              private key هرگز نباید به browser منتقل شود.
            </div>
          </div>
        </div>
      </section>

      {selected && (
        <div
          style={{
            position: 'fixed',
            inset: 0,
            zIndex: 100,
            background: 'rgba(0,0,0,.68)',
            display: 'grid',
            placeItems: 'center',
            padding: 20,
          }}
        >
          <div
            className="card panel"
            style={{ width: 'min(520px,100%)' }}
          >
            <div className="split">
              <h2 style={{ margin: 0 }}>Bind & Issue</h2>

              <button
                className="btn btn-ghost"
                onClick={() => setSelected(null)}
                disabled={busy}
              >
                ×
              </button>
            </div>

            <div className="spaced">
              <div className="order-meta">
                <div className="meta">
                  <span>Customer</span>
                  <b>{selected.customer}</b>
                </div>

                <div className="meta">
                  <span>Branch</span>
                  <b>branch-demo-01</b>
                </div>

                <div className="meta">
                  <span>Industry</span>
                  <b>industrial-monitoring</b>
                </div>

                <div className="meta">
                  <span>Hardware</span>
                  <b className="mono">
                    HW-BIND-DEMO-01
                  </b>
                </div>
              </div>

              <div className="notice spaced">
                این مرحله Demo است. private key در browser قرار
                نمی‌گیرد. اتصال واقعی باید به سرویس امن Horizon Core
                انجام شود.
              </div>

              <button
                className="btn btn-primary spaced"
                style={{ width: '100%' }}
                onClick={issue}
                disabled={busy}
              >
                {busy ? 'در حال صدور...' : 'ایجاد و ثبت لایسنس'}
              </button>
            </div>
          </div>
        </div>
      )}
    </main>
  );
}
PAGEEOF

echo "[7/8] بررسی TypeScript و build..."

if [ -f package-lock.json ]; then
  npm install --ignore-scripts
elif [ -f pnpm-lock.yaml ]; then
  corepack pnpm install --ignore-scripts
elif [ -f yarn.lock ]; then
  yarn install --ignore-scripts
else
  npm install --ignore-scripts
fi

echo
echo "========================================"
echo " TYPE CHECK / BUILD"
echo "========================================"

if npm run | grep -q "typecheck"; then
  npm run typecheck
fi

if npm run | grep -q "lint"; then
  npm run lint
fi

npm run build

echo
echo "========================================"
echo " SUCCESS"
echo "========================================"
echo
echo "Horizon Angel اصلاح و Build شد."
echo
echo "برای اجرای محلی:"
echo "  npm run dev"
echo
echo "سپس:"
echo "  http://localhost:3000/angel"
echo
