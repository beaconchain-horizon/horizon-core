'use client';
import Link from 'next/link';
import {products} from '../../lib/demo-data';
import {useEffect,useState} from 'react';

export default function Products(){
 const [cart,setCart]=useState<any[]>([]);
 useEffect(()=>{try{setCart(JSON.parse(localStorage.getItem('hz_cart')||'[]'))}catch{}} ,[]);
 const add=(id:string)=>{const next=[...cart,{id,quantity:1}];setCart(next);localStorage.setItem('hz_cart',JSON.stringify(next));alert('محصول به سبد اضافه شد.')};
 return <main><section className="page-head"><div className="container"><h1>محصولات Horizon</h1><p>صفحه عمومی فروشگاه. محصول مانیتورینگ صنعتی یک بخش اصلی است و دمو برای مشاهده قابلیت‌ها در دسترس است.</p></div></section><section className="section"><div className="container grid-3">{products.map(p=><div className="card product-card" key={p.id}><div className="product-top"><div className="icon-box">{p.icon}</div><span className="badge">{p.tag}</span></div><h3>{p.name}</h3><p>{p.desc}</p><div className="price">{p.price}</div><div style={{display:'grid',gap:8}}>{p.id==='monitoring'&&<Link className="btn" href="/monitor-demo">دموی عمومی</Link>}<button className="btn btn-primary" onClick={()=>add(p.id)}>افزودن به سفارش</button></div></div>)}</div><div className="container spaced"><div className="split"><span className="muted small">سبد فعلی: {cart.length} قلم</span><Link href="/cart" className="btn">رفتن به سبد</Link></div></div></section></main>
}
