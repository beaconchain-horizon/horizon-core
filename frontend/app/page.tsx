import Link from 'next/link';
import { products } from '../lib/demo-data';

export default function Home(){return <main>
<section className="hero"><div className="container hero-grid">
<div>
<span className="eyebrow"><span className="dot"/> زیرساخت اختصاصی · فروشگاه عمومی + کنترل خصوصی</span>
<h1><span className="gradient-text">Horizon</span><br/>برای شبکه‌هایی که باید قابل کنترل بمانند.</h1>
<p className="lead">یک ویترین عمومی تمیز برای خرید و سفارش، در کنار Horizon Angel برای مدیریت خصوصی مشتری‌ها، شعبه‌ها، صنایع، Agentها و صدور لایسنس اختصاصی.</p>
<div className="hero-actions"><Link href="/products" className="btn btn-primary">مشاهده محصولات</Link><Link href="/monitor-demo" className="btn">مشاهده دموی مانیتورینگ</Link><Link href="/about" className="btn btn-ghost">آشنایی با Horizon</Link></div>
</div>
<div className="hero-card"><div className="mini-grid">
<div className="metric"><div className="label">معماری</div><div className="value">Air-Gap</div><div className="sub">آماده برای استقرار ایزوله</div></div>
<div className="metric"><div className="label">امضای لایسنس</div><div className="value">ECDSA</div><div className="sub">P-256 · سمت امن مالک</div></div>
<div className="metric"><div className="label">اثبات یکپارچگی</div><div className="value">Merkle</div><div className="sub">ریشه قابل بررسی</div></div>
<div className="metric"><div className="label">کنترل خصوصی</div><div className="value">Angel SOC</div><div className="sub">مخصوص مالک سیستم</div></div>
</div><div className="hr"/><div className="small muted">دامنه نهایی بعداً در <span className="mono">.env.local</span> تنظیم می‌شود.</div></div>
</div></section>

<section className="section"><div className="container"><div className="section-title"><div><h2>محصول اصلی: مانیتورینگ صنعتی</h2><p>برای سایت‌های نفت، گاز، پتروشیمی، فولاد و دیگر محیط‌های حساس.</p></div><Link href="/products/monitoring" className="btn btn-ghost">جزئیات محصول</Link></div>
<div className="two-col"><div className="card" style={{padding:22}}><div className="feature-list">
<div className="feature"><div className="feature-icon">◎</div><div><b>سنسور و وضعیت سایت</b><span>نمایش داده‌های سنسور، سلامت gateway و رخدادهای مهم.</span></div></div>
<div className="feature"><div className="feature-icon">⌁</div><div><b>آماده برای Air-Gap</b><span>دموی آنلاین برای معرفی قابلیت‌ها؛ استقرار واقعی می‌تواند ایزوله و کنترل‌شده باشد.</span></div></div>
<div className="feature"><div className="feature-icon">⌕</div><div><b>لایسنس اختصاصی</b><span>هر سفارش می‌تواند به مشتری، شعبه، صنعت و hardware identity مشخص متصل شود.</span></div></div>
</div></div><div className="industry-visual"><div className="grid-lines"/><div className="monitor-screen"><div className="screen-head"><span>Horizon Industrial Monitor</span><span className="ok-text">● LIVE</span></div><div className="signal-row"><div className="signal"><small>Pressure</small><div className="v">84.2</div></div><div className="signal"><small>Temp</small><div className="v">72.8°C</div></div><div className="signal"><small>Alerts</small><div className="v">02</div></div></div><div className="spark"><svg viewBox="0 0 480 120" preserveAspectRatio="none"><path d="M0 86 C40 75 45 92 78 61 S120 73 145 49 178 69 208 31 240 45 267 52 300 38 323 44 350 22 378 34 405 26 430 41 455 18 480 23" fill="none" stroke="currentColor" strokeWidth="3"/></svg></div></div></div></div>
</div></section>

<section className="section"><div className="container"><div className="section-title"><div><h2>سه مسیر ساده</h2><p>برای مشتری، برای صاحب سیستم، برای لایسنس.</p></div></div><div className="grid-3">{products.map(p=><div className="card product-card" key={p.id}><div className="product-top"><div className="icon-box">{p.icon}</div><span className="badge">{p.tag}</span></div><h3>{p.name}</h3><p>{p.desc}</p><div className="price">{p.price}</div><Link href={p.id==='monitoring'?'/products/monitoring':'/products'} className="btn btn-primary" style={{width:'100%'}}>مشاهده</Link></div>)}</div></div></section>

<section className="section"><div className="container card" style={{padding:24}}><div className="split"><div><h2 style={{margin:'0 0 8px'}}>منوی روشن، بدون شلوغی</h2><div className="muted" style={{fontSize:13}}>فروشگاه عمومی، دموی مانیتورینگ، اسناد فنی، پنل مشتری و Horizon Angel از هم تفکیک شده‌اند.</div></div><Link href="/documents" className="btn">اسناد و گواهی‌ها</Link></div></div></section>
</main>}
