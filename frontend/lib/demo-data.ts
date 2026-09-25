export const products = [
  {id:'monitoring',icon:'◉',name:'Horizon Industrial Monitoring',tag:'صنعت',desc:'مانیتورینگ سنسورها، هشدارها و وضعیت سایت با معماری آماده برای محیط‌های ایزوله.',price:'تماس / سفارش',featured:true},
  {id:'banking',icon:'◆',name:'Horizon Core Banking',tag:'بانکی',desc:'زیرساخت دفترکل و عملیات امضاشده برای استقرارهای سازمانی و کنترل‌شده.',price:'توافقی'},
  {id:'license',icon:'▣',name:'Secure License',tag:'لایسنس',desc:'لایسنس اختصاصی با اتصال به مشتری، شعبه، صنعت و hardware identity.',price:'بر اساس سفارش'},
];

export const demoOrders = [
  {id:'ORD-DEMO-1842',customer:'Sample Oil Group',product:'Horizon Industrial Monitoring',qty:12,status:'pending_issuance',created:'امروز · 09:42'},
  {id:'ORD-DEMO-1836',customer:'Sample Bank',product:'Horizon Core Banking',qty:4,status:'issued',created:'دیروز · 17:10'},
  {id:'ORD-DEMO-1811',customer:'Sample Steel Co.',product:'Horizon Industrial Monitoring',qty:8,status:'awaiting_payment',created:'2 روز قبل'},
];

export const agents = Array.from({length:144},(_,i)=>({id:i+1,name:`HZ-Agent-${String(i+1).padStart(3,'0')}`,status:i%11===0?'offline':i%17===0?'warning':'online',site:`site-${String((i%18)+1).padStart(2,'0')}`}));

export const customers=[
 {id:'bank_mellat',name:'بانک ملت',type:'بانک',branches:1,license:'L-MLT-01',status:'active'},
 {id:'sample-oil',name:'گروه نمونه نفت',type:'صنعت نفت',branches:6,license:'L-OIL-06',status:'active'},
 {id:'sample-steel',name:'نمونه فولاد',type:'فولاد',branches:4,license:'L-STL-04',status:'active'},
 {id:'sample-gas',name:'نمونه گاز',type:'گاز',branches:3,license:'L-GAS-03',status:'grace'},
];

export const certDocs=[
 {title:'Technical Certificate',desc:'گواهی فنی نسخه 3.0.0 شامل معماری، اجزای رمزنگاری و نتایج تست‌های ثبت‌شده در repository.',href:'/documents#certificate'},
 {title:'Air-Gap Architecture',desc:'دو فاز Air-Gap Runtime و Controlled Data Transfer و اصول تبادل داده امضاشده.',href:'/documents#airgap'},
 {title:'Security Roadmap',desc:'مسیر امنیتی پروژه و نقاط بررسی برای توسعه و سخت‌سازی.',href:'/documents#security'},
];
