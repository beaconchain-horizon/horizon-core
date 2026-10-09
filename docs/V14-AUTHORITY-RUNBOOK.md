# Horizon Core v1.4 — Authority and license runbook

این راهنما برای راه‌اندازی تازهٔ Horizon Core v1.4 است. لایسنس‌های تستی قدیمی عمداً سازگار نگه داشته نمی‌شوند. نسخهٔ محصول روی v1.4 و زنجیرهٔ ابزار روی Go 1.24.x ثابت می‌ماند.

## 1. پیش‌نیازها

در Git Bash ویندوز، داخل ریشهٔ مخزن:
- Go 1.24.x
- Docker Desktop با Docker Compose v2
- `curl` و `openssl` (در Git Bash موجود باشند)

## 2. راه‌اندازی یک‌دست

ابتدا تغییرات شاخهٔ کاری را دریافت کن و مطمئن شو روی شاخهٔ یکپارچه هستی:

```bash
git fetch origin
git switch feat/v14-unified-authority-license-20261009
git pull --ff-only origin feat/v14-unified-authority-license-20261009
```

سپس:

```bash
bash scripts/bootstrap-horizon-v14.sh
```

اسکریپت قبل از تغییر کلیدها و `.env` از موارد موجود نسخهٔ پشتیبان زمان‌دار می‌سازد، شناسهٔ Authority جدید secp256k1 را تولید می‌کند، تنظیمات لازم را می‌نویسد، Compose را اعتبارسنجی می‌کند و سرویس‌ها را بالا می‌آورد. این اسکریپت به `main` چیزی merge نمی‌کند و به Liara deploy نمی‌کند.

## 3. بررسی سرویس‌ها

```bash
curl -fsS http://127.0.0.1:8081/health
curl -fsS http://127.0.0.1:8080/api/v1/health
docker compose ps
```

برای مشاهدهٔ شناسهٔ سخت‌افزاری‌ای که خود Switch اعمال می‌کند، مقدار `ADMIN_TOKEN` را از فایل محلی `.env` بخوان و درخواست محافظت‌شده بفرست:

```bash
ADMIN_TOKEN="$(sed -n 's/^ADMIN_TOKEN=//p' .env)"
curl -fsS -H "X-Admin-Token: $ADMIN_TOKEN" \
  http://127.0.0.1:8081/api/v1/license/hardware-id
```

در استقرار ابری که شناسهٔ DMI در کانتینر قابل خواندن نیست، مقدار پایدار `HORIZON_HARDWARE_ID` را در تنظیمات secret/config همان سرویس Switch تنظیم کن. اگر Switch نتواند HWID تعیین کند، اعتبارسنجی باید fail-closed شود؛ از HWID دلخواه یا خالی استفاده نکن.

## 4. صدور یک لایسنس تازه

شناسهٔ سخت‌افزاری مرحلهٔ قبل را در `HWID` قرار بده. فرمان زیر فقط یک نمونهٔ قراردادی برای آزمون محلی ایجاد می‌کند، نه مشتری واقعی:

```bash
ADMIN_TOKEN="$(sed -n 's/^ADMIN_TOKEN=//p' .env)"
API_KEY="$(sed -n 's/^API_KEY=//p' .env)"
HWID="PUT_SWITCH_HARDWARE_ID_HERE"

curl -fsS -X POST http://127.0.0.1:8080/api/v1/license/generate \
  -H "X-API-Key: $API_KEY" \
  -H "Content-Type: application/json" \
  --data "{\"product_id\":\"horizon-core\",\"user_id\":\"local-validation-tenant\",\"volume\":1000,\"duration\":30,\"hardware_id\":\"$HWID\"}"
```

پاسخ شامل `license` (ساختار امضاشدهٔ واحد) است. برای ذخیرهٔ خود envelope جهت بررسی مستقل، در صورت نصب بودن `jq`:

```bash
# پاسخ curl را در response.json ذخیره کرده باشی:
jq '.license' response.json > license.json
```

## 5. اعتبارسنجی مستقل و آفلاین

این فرمان فقط از لایسنس، کلید عمومی مورد اعتماد و زمان سیستم استفاده می‌کند؛ به پاسخ تأییدی Backend وابسته نیست:

```bash
go run ./cmd/licenseverify \
  -license license.json \
  -public-key secrets/v1.4/issuer-public.hex \
  -hardware-id "$HWID"
```

کلید خصوصی در `secrets/v1.4/issuer-private.pem` قرار می‌گیرد و نباید commit یا به مرورگر ارسال شود. فایل `secrets/` در `.gitignore` قرار دارد. کلید عمومی را مستقل از فایل لایسنس و از مسیر قابل اعتماد provision کن؛ کلید عمومی داخل لایسنس به‌تنهایی مرجع اعتماد نیست.

## 6. آزمون زندهٔ جدید v1.4 (فقط محیط محلی)

این تست عمداً فقط URLهای loopback را می‌پذیرد تا ناخواسته روی سرویس تولیدی لایسنس آزمایشی نسازد. پس از اجرای راه‌اندازی تازه و healthy شدن هر دو سرویس، در Git Bash اجرا کن:

```bash
export API_KEY="$(sed -n 's/^API_KEY=//p' .env)"
export ADMIN_TOKEN="$(sed -n 's/^ADMIN_TOKEN=//p' .env)"
go run ./cmd/license-e2e
unset API_KEY ADMIN_TOKEN
```

تست از Switch می‌خواند و HWID، از Backend لایسنس واقعی v1.4 می‌گیرد، آن را به مسیر جدید `/api/v1/license/check` می‌فرستد و باید چهار حالت را بررسی کند: لایسنس سالم پذیرفته شود؛ امضای تغییرکرده رد شود؛ statement تغییرکرده رد شود؛ و HWID ناسازگار صریحاً رد شود. خروجی امضا و توکن را چاپ نمی‌کند. این تست یک لایسنس اعتبارسنجی در دیتابیس محلی Backend و Switch باقی می‌گذارد؛ فقط در محیط محلی/تستی اجرا کن. تست قدیمی pipe-delimited برای ممیزی تاریخی به مسیر `/api/v1/license/check-legacy` منتقل شده است و اثبات پروتکل جدید محسوب نمی‌شود.

## 7. توقف/عیب‌یابی

```bash
docker compose ps
docker compose logs --tail=100 backend switch postgres
```

خطای HWID، امضا، ریشهٔ مرکل، انقضا یا کلید Authority باید لایسنس را رد کند، نه این‌که اعتبارسنجی را خاموش کند. برای عیب‌یابی تنظیمات را باز کن؛ فایل‌های کلید خصوصی را در Issue/PR/گفت‌وگو کپی نکن.

## مرز مهم اثبات

در این نسخه، ریشهٔ مرکل فیلدهای immutable لایسنس را پوشش می‌دهد و امضای Authority اصالت statement را بررسی می‌کند. اثبات inclusion این لایسنس در یک رجیستری سراسری append-only با checkpoint/root منتشرشده، هنوز کار مستقلی است و نباید صرفاً به‌دلیل وجود این root ادعای آن مطرح شود.
