'use client';

import type { FormEvent } from 'react';
import { useEffect, useState } from 'react';

type Summary = {
  ok?: boolean;
  source?: string;
  status?: string;
  version?: string;
  message?: string;
  stats?: Record<string, unknown> | null;
};

type IssueResult = {
  ok?: boolean;
  license_id?: string;
  message?: string;
  error?: string;
  license?: unknown;
};

export default function Angel() {
  const [summary, setSummary] = useState<Summary | null>(null);
  const [summaryMessage, setSummaryMessage] = useState('در حال اتصال به Switch...');
  const [customerID, setCustomerID] = useState('');
  const [branchID, setBranchID] = useState('');
  const [productID, setProductID] = useState('industrial-monitoring');
  const [hardwareID, setHardwareID] = useState('');
  const [volume, setVolume] = useState('1000');
  const [duration, setDuration] = useState('30');
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState('');
  const [issued, setIssued] = useState<IssueResult | null>(null);

  useEffect(() => {
    let mounted = true;
    fetch('/api/angel/summary', { cache: 'no-store' })
      .then(async (r) => {
        const data = await r.json().catch(() => ({})) as Summary;
        if (r.status === 401) {
          window.location.href = '/angel/login';
          return;
        }
        if (!r.ok) throw new Error(data.message || 'ارتباط با Switch برقرار نشد.');
        if (mounted) {
          setSummary(data);
          setSummaryMessage('اتصال زنده برقرار است.');
        }
      })
      .catch((e: unknown) => {
        if (mounted) {
          setSummary(null);
          setSummaryMessage(e instanceof Error ? e.message : 'وضعیت زنده در دسترس نیست.');
        }
      });
    return () => { mounted = false; };
  }, []);

  async function issueLicense(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (busy) return;
    setBusy(true);
    setError('');
    setIssued(null);
    try {
      const response = await fetch('/api/angel/issue-license', {
        method: 'POST',
        headers: { 'content-type': 'application/json' },
        cache: 'no-store',
        body: JSON.stringify({
          customer_id: customerID.trim(),
          branch_id: branchID.trim(),
          product_id: productID.trim(),
          hardware_id: hardwareID.trim(),
          volume: Number(volume),
          duration: Number(duration),
        }),
      });
      const data = await response.json().catch(() => ({})) as IssueResult;
      if (response.status === 401) {
        window.location.href = '/angel/login';
        return;
      }
      if (!response.ok || !data.ok || !data.license) {
        throw new Error(data.message || data.error || 'صدور امن لایسنس انجام نشد.');
      }
      setIssued(data);
    } catch (e: unknown) {
      setError(e instanceof Error ? e.message : 'خطای نامشخص در صدور لایسنس');
    } finally {
      setBusy(false);
    }
  }

  async function logout() {
    await fetch('/api/angel/login', { method: 'DELETE' }).catch(() => undefined);
    window.location.href = '/angel/login';
  }

  return (
    <main>
      <section className="page-head">
        <div className="container">
          <div className="split">
            <div>
              <span className="eyebrow"><span className="dot" /> PRIVATE OWNER CONSOLE</span>
              <h1>Horizon Angel · Authority</h1>
              <p className="muted">
                صدور لایسنس v1.4 از Horizon Authority؛ کلید خصوصی فقط در Backend نگه‌داری می‌شود.
              </p>
            </div>
            <button className="btn btn-danger" onClick={logout}>خروج امن</button>
          </div>
        </div>
      </section>

      <section className="section">
        <div className="container">
          <div className="card panel">
            <span className="eyebrow">LIVE SWITCH STATUS</span>
            <h2>{summary?.status === 'online' ? 'Switch متصل است' : 'وضعیت سرویس'}</h2>
            <p className={summary?.status === 'online' ? 'ok-text' : 'warn-text'}>{summaryMessage}</p>
            {summary?.status === 'online' && (
              <p className="muted small">
                سرویس: {summary.source || 'horizon-switch'} · نسخه: {summary.version || 'نامشخص'}
              </p>
            )}
            <p className="muted small">
              تعداد Agent و سفارش تا زمانی که منبع عملیاتی واقعی متصل نباشد نمایش داده نمی‌شود؛ دادهٔ ساختگی به‌عنوان وضعیت زنده نشان داده نمی‌شود.
            </p>
          </div>
        </div>
      </section>

      <section className="section">
        <div className="container">
          <div className="card panel">
            <span className="eyebrow">AUTHORITY · ECDSA SECP256K1</span>
            <h2>صدور لایسنس جدید</h2>
            <p className="muted small">
              HWID را از مسیر محافظت‌شدهٔ Switch یعنی <code>/api/v1/license/hardware-id</code> بگیر. HWID خالی یا ساختگی استفاده نکن.
            </p>
            <form onSubmit={issueLicense} className="license-form">
              <div className="field">
                <label htmlFor="customer-id">شناسه مشتری / Tenant</label>
                <input id="customer-id" value={customerID} onChange={(e) => setCustomerID(e.target.value)} required maxLength={200} autoComplete="off" />
              </div>
              <div className="field">
                <label htmlFor="branch-id">شناسه شعبه / سایت</label>
                <input id="branch-id" value={branchID} onChange={(e) => setBranchID(e.target.value)} required maxLength={200} autoComplete="off" />
              </div>
              <div className="field">
                <label htmlFor="product-id">شناسه محصول</label>
                <input id="product-id" value={productID} onChange={(e) => setProductID(e.target.value)} required maxLength={200} autoComplete="off" />
              </div>
              <div className="field">
                <label htmlFor="hardware-id">Hardware ID (HWID)</label>
                <input id="hardware-id" value={hardwareID} onChange={(e) => setHardwareID(e.target.value)} required maxLength={512} autoComplete="off" />
              </div>
              <div className="field">
                <label htmlFor="volume">ظرفیت / Volume</label>
                <input id="volume" type="number" min={1} max={10000000} value={volume} onChange={(e) => setVolume(e.target.value)} required />
              </div>
              <div className="field">
                <label htmlFor="duration">مدت اعتبار (روز)</label>
                <input id="duration" type="number" min={1} max={3650} value={duration} onChange={(e) => setDuration(e.target.value)} required />
              </div>
              <button className="btn btn-primary" type="submit" disabled={busy}>
                {busy ? 'در حال صدور امن…' : 'صدور و امضای لایسنس'}
              </button>
            </form>
            {error && <div className="notice danger-text" role="alert">{error}</div>}
            {issued && Boolean(issued.license) && (
              <div className="spaced">
                <div className="notice ok-text">لایسنس جدید با مرجع Authority صادر شد: <b>{issued.license_id}</b></div>
                <p className="muted small">JSON زیر envelope امضاشده است؛ کلید خصوصی یا API key در آن قرار ندارد.</p>
                <pre className="license-output">{JSON.stringify(issued.license, null, 2)}</pre>
                <button className="btn btn-secondary" type="button" onClick={() => {
                  void navigator.clipboard?.writeText(JSON.stringify(issued.license, null, 2));
                }}>کپی JSON لایسنس</button>
              </div>
            )}
          </div>
        </div>
      </section>
    </main>
  );
}
