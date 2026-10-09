# Horizon Core — Static Documentation and Demo

این پوشه شامل سایت استاتیک HTML/CSS/JavaScript پروژه است و با workflow موجود در `.github/workflows/static.yml` به GitHub Pages منتشر می‌شود. برای این سایت نصب Node.js یا ساخت Next.js لازم نیست.

## صفحات اصلی

- `index.html` — داشبورد و صفحهٔ اصلی
- `dashboard.html` و `admin.html` — داشبورد و پنل مدیریت
- `industry.html` — نمای صنعتی
- `products.html`، `cart.html` و `checkout.html` — صفحات محصولات و سفارش
- `license.html` — صفحهٔ لایسنس
- `login.html` و `register.html` — ورود و ثبت‌نام
- `bank.html`، `central-bank.html`، `bank-node-panel.html` و `bank-saderat-dashboard.html` — نماهای بانکی
- `horizon-dex.html` و `news-media.html` — صفحات اطلاعاتی
- `neshan.html`، `neshan-check.html` و `neshan-diag.html` — صفحات بررسی نقشه
- `certificates/COMMERCIAL-RECORD.html` — سند تجاری

## فایل‌های مشترک

- `config.js` — تنظیمات endpointهای فرانت استاتیک
- `js/api.js` — ارتباط API
- `js/main.js`، `js/cart.js` و `js/demo-data.js` — رفتارهای سمت مرورگر
- `css/style.css` — استایل
- `sw.js` — Service Worker
- `offline-storage.js` و `clients.js` — ذخیره‌سازی آفلاین و اطلاعات کلاینت

## انتشار

محتوای `docs/` مستقیماً منتشر می‌شود. فایل‌های Next.js در این مخزن نگهداری نمی‌شوند.
