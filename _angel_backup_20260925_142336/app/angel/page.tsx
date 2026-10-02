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
