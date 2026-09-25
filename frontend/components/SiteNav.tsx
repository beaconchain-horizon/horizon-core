'use client';
import Link from 'next/link';
import { usePathname } from 'next/navigation';

export function SiteNav(){
  const path = usePathname();
  const links = [
    ['/', 'خانه'],['/products','محصولات'],['/monitor-demo','دموی مانیتورینگ'],['/about','درباره Horizon'],['/documents','اسناد و گواهی'],['/customer','پنل مشتری']
  ];
  return <header className="topbar"><div className="container nav">
    <Link href="/" className="brand"><span className="brand-mark">H</span><span>Horizon<small>Secure infrastructure</small></span></Link>
    <nav className="navlinks">{links.map(([href,label])=><Link key={href} href={href} className={path===href?'active':''}>{label}</Link>)}</nav>
    <div className="nav-cta"><Link href="/cart" className="btn btn-ghost">سبد</Link><Link href="/angel" className="btn btn-primary">Angel SOC</Link></div>
  </div></header>
}
