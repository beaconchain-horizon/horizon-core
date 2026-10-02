'use client';

import {
  angelAgents,
  angelCustomers,
  angelDemoOrders,
} from '../../lib/angel-demo-data';

import { useEffect, useMemo, useState } from 'react';

type Order = (typeof angelDemoOrders)[number];

type LiveSummary = {
  agents: number;
  online: number;
  warning: number;
  offline: number;
  source?: string;
  timestamp?: string;
};

export default function Angel() {
  const [orders, setOrders] = useState<Order[]>(angelDemoOrders);
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
        if (!r.ok) {
          throw new Error('summary request failed');
        }

        return r.json();
      })
      .then((data: LiveSummary) => {
        if (mounted) {
          setLive(data);
        }
      })
      .catch(() => {
        if (mounted) {
          setLive(null);
        }
      });

    return () => {
      mounted = false;
    };
  }, []);

  const stats = useMemo(
    () => ({
      total: live?.agents ?? angelAgents.length,
      online:
        live?.online ??
        angelAgents.filter((a) => a.status === 'online').length,
      warning:
        live?.warning ??
        angelAgents.filter((a) => a.status === 'warning').length,
      offline:
        live?.offline ??
        angelAgents.filter((a) => a.status === 'offline').length,
    }),
    [live]
  );

  const issue = async () => {
    if (!selected || busy) {
      return;
    }

    setBusy(true);
    setMsg('');

    try {
      const response = await fetch(
        '/api/angel/issue-license',
        {
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
        }
      );

      const data = await response.json();

      if (!response.ok) {
        throw new Error(
          data?.message || 'صدور لایسنس ناموفق بود.'
        );
      }

      setMsg(`صدور موفق: ${data.license_id}`);

      setOrders((current) =>
        current.map((order) =>
          order.id === selected.id
            ? { ...order, status: 'issued' }
            : order
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
                کنترل مرکزی برای Agentها، مشتری‌ها، سفارش‌ها و
                صدور لایسنس. داده‌های عملیاتی باید از API امن
                Horizon بیایند.
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
                وضعیت Agentها از API خوانده می‌شود؛ در حالت فعلی
                داده‌ها Demo هستند.
              </p>
            </div>
          </div>

          <div className="agent-grid">
            {angelAgents.map((agent) => (
              <div className="agent" key={agent.id}>
                <b>{agent.name}</b>
                <small>{agent.site}</small>

                <div
                  className={`state ${
                    agent.status === 'offline'
                      ? 'off'
                      : agent.status === 'warning'
                        ? 'warn'
                        : ''
                  }`}
                >
                  <i /> {agent.status}
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
              {orders.map((order) => (
                <div
                  className="order-card"
                  key={order.id}
                  style={{
                    border: '1px solid var(--line)',
                    borderRadius: 16,
                    marginBottom: 10,
                    padding: 14,
                  }}
                >
                  <div className="split">
                    <div>
                      <b>{order.customer}</b>

                      <div className="muted small">
                        {order.product} · {order.qty} license
                      </div>
                    </div>

                    <span
                      className={`badge ${
                        order.status === 'issued'
                          ? 'ok'
                          : order.status === 'pending_issuance'
                            ? 'warn'
                            : ''
                      }`}
                    >
                      {order.status}
                    </span>
                  </div>

                  <div className="split">
                    <span className="muted small mono">
                      {order.id}
                    </span>

                    {order.status === 'pending_issuance' && (
                      <button
                        className="btn btn-primary"
                        onClick={() => setSelected(order)}
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

            {angelCustomers.map((customer) => (
              <div
                className="status-row"
                key={customer.id}
              >
                <div>
                  <b>{customer.name}</b>

                  <div className="muted small">
                    {customer.type} · {customer.branches} branch
                  </div>
                </div>

                <span
                  className={
                    customer.status === 'active'
                      ? 'ok-text'
                      : 'warn-text'
                  }
                >
                  {customer.status}
                </span>
              </div>
            ))}

            <div className="hr" />

            <div className="notice">
              hardware_id باید در binding واقعی لایسنس قرار
              بگیرد و private key هرگز نباید به browser منتقل
              شود.
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
                نمی‌گیرد. اتصال واقعی باید به سرویس امن Horizon
                Core انجام شود.
              </div>

              <button
                className="btn btn-primary spaced"
                style={{ width: '100%' }}
                onClick={issue}
                disabled={busy}
              >
                {busy
                  ? 'در حال صدور...'
                  : 'ایجاد و ثبت لایسنس'}
              </button>
            </div>
          </div>
        </div>
      )}
    </main>
  );
}
