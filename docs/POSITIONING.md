# Horizon Core v1.4 — Deployment Positioning

**Date**: 2026-10-11
**Audience**: Sarcheshmeh (industrial) · Judiciary IT / Khorasan (regulatory)

---

## خلاصه‌ی مدیریتی

Horizon Core یک **لایه‌ی اثبات‌پذیری و حاکمیت** است که روی زیرساخت موجود نصب می‌شود.
جایگزین Core Banking نمی‌شود، جایگزین SCADA نمی‌شود. کاری که می‌کند:

1. **امضای دیجیتال هر عملیات** (secp256k1)
2. **زنجیر تغییرناپذیر شواهد** (Merkle RFC 6962)
3. **کنترل لایسنس و سیاست** (Policy Enforcement Gateway)
4. **کار در حالت air-gap**

هدف: وقتی رگولاتور یا حسابرس می‌پرسد «این عدد از کجا آمده؟»،
شما یک proof رمزنگاری‌شده نشان می‌دهید، نه یک log قابل ویرایش.

---

## بخش الف — جایگاه در لایه‌های بانکداری

```
L7 │ کانال‌های مشتری   │ اینترنت‌بانک · موبایل · ATM · POS
L6 │ واسط / API Gateway │ REST · ISO 8583 · SOAP
L5 │ سوییچ پرداخت       │ RTGS · ACH · شتاب · کارت
L4 │ Core Banking (CBS) │ حساب · تراکنش · تسویه
L3 │ Treasury/ریسک      │ نقدینگی · اعتبار · نرخ
L2 │ انطباق و AML       │ AML · KYC · Fraud · رگولاتور
L1 │ حسابرسی و شواهد    │ Audit · Evidence · Reporting  ← Horizon
L0 │ زیرساخت            │ DB · Storage · شبکه · HSM
```

### Horizon کجا می‌نشیند؟

**بین L1 و L2** — لایه‌ی شواهد و حاکمیت، با hooks به L5 و L4.

| تماس با لایه | نحوه‌ی اتصال | جهت |
|---|---|---|
| L5 (سوییچ پرداخت) | امضای هر پیام ISO/RTGS | outbound |
| L4 (CBS) | anchor هر تراکنش تسویه | inbound |
| L3 (Treasury) | register گزارش‌های روزانه | inbound |
| L2 (AML/Fraud) | ارائه proof به رگولاتور | outbound |
| L1 (Audit) | ذخیره‌ی زنجیر تغییرناپذیر | بومی |
| L0 (زیرساخت) | مستقل — DB + Storage موجود | passive |

### چه چیزی را حل می‌کند؟

1. اثبات اصالت تراکنش به رگولاتور
2. جلوگیری از دستکاری log توسط insider
3. Dual Control رمزنگاری‌شده
4. گره‌خوردگی لایسنس به سخت‌افزار

### چه چیزی نیست؟

- ❌ جایگزین Core Banking نیست
- ❌ جایگزین سوییچ پرداخت نیست
- ❌ اجماع توزیع‌شده ندارد
- ❌ HSM سخت‌افزاری ندارد (نسخه‌ی بعدی)

---

## بخش ب — جایگاه در لایه‌های صنعت (Purdue ISA-95)

```
L5 │ Cloud / Enterprise SaaS │ BI · ERP ابری · گزارش رگولاتور
L4 │ Enterprise (IT)         │ ERP · MES · Historian مرکزی
L3 │ Operations (OT)         │ Historian · MES · Batch
L2 │ Supervision             │ SCADA · HMI · Alarm
L2.5│ ← Horizon (Evidence)  │ anchor · Merkle · air-gap
L1 │ Control                 │ PLC · RTU · DCS
L0 │ Field                   │ سنسور · عملگر · فیلدباس
```

### Horizon کجا می‌نشیند؟

**L2.5** — لایه‌ی اثبات داده‌ی صنعتی.

| تماس با لایه | نحوه‌ی اتصال | جهت |
|---|---|---|
| L2 (SCADA/HMI) | دریافت خوانش (OPC-UA · Modbus) | inbound |
| L2.5 (این سیستم) | anchor هر خوانش | بومی |
| L3 (Historian) | ارسال hash | outbound |
| L3 (MES) | اتصال به batch records | outbound |
| L4 (ERP) | گزارش دوره‌ای | outbound |
| Air-Gap | جداسازی کامل | بومی |

### چه چیزی را حل می‌کند؟

1. اثبات اصالت داده‌ی SCADA به رگولاتور
2. جلوگیری از دستکاری داده‌ی تاریخی
3. کار در مناطق دورافتاده بدون اینترنت
4. پاسخ رمزنگاری به «این عدد از کجا آمده؟»

### مثال سرچشمه

| نیاز | راه‌حل Horizon |
|---|---|
| ثبت فشار/دما/جریان | anchor هر ۳۰ ثانیه |
| اثبات به سازمان محیط زیست | Merkle proof |
| جداسازی از IT | air-gap mode |
| حسابرسی | زنجیر تغییرناپذیر |

---

## بخش ج — وضعیت امنیتی

### تست‌شده و کار می‌کند

| قابلیت | اثبات | تست |
|---|---|---|
| امضای secp256k1 | ۱۲۸ hex | ✅ crypto/secp256k1_test.go |
| فرمت R||S 64-byte | بدون DER/P-256 | ✅ reject tests |
| RFC 6979 deterministic | دو امضا یکسان | ✅ TestSign_IsDeterministic |
| rejection تغییر بایت امضا | tamper → reject | ✅ TestTamper_* |
| rejection تغییر payload | tamper → reject | ✅ TestTamper_RejectsPayloadChange |
| HWID binding | mismatch → reject | ✅ TestHWID_Policy |
| expiry | خارج از بازه → reject | ✅ TestExpiry_Boundaries |
| Merkle inclusion | tamper leaf → reject | ✅ TestProof_RejectsTamperedLeaf |
| کار بدون اینترنت | امضا USB + verify | ✅ E2E |

### Roadmap (صادقانه)

| قابلیت | وضعیت | ETA |
|---|---|---|
| Multi-node consensus | طراحی | Q1 2027 |
| HSM | انتخاب vendor | Q4 2026 |
| ISO 27001 | مستندسازی | Q2 2027 |
| FIPS 186-5 | **درخواست نشده** | Q3 2027 |

> **نکته برای دادگستری:** ادعای FIPS نداریم. secp256k1 استاندارد SEC 2 است
> (Bitcoin/Ethereum) ولی تحت FIPS 186-5 تأیید نشده.

### جایگاه امنیتی — چه چیزی محافظت می‌شود

| تهدید | محافظت | لایه |
|---|---|---|
| دستکاری log | زنجیر تغییرناپذیر | L1 |
| جعل امضا | secp256k1 | L1 |
| replay | nonce + timestamp | L2 |
| SSRF | validation ۳۹/۴۰ | L2 |
| DoS | rate + connection limit | L1 |
| دسترسی غیرمجاز API | fail-closed | L3 |
| سرقت کلید خصوصی | USB offline | L3 |
| نصب سخت‌افزار اشتباه | HWID binding | L3 |
| نشت وابستگی | Dependabot + CodeQL | L4 |
| دسترسی خارجی | air-gap | L5 |

---

## پیشنهاد ارائه

### برای سرچشمه (۱۵ دقیقه)

1. مسئله: چه کسی ثابت می‌کند داده دست‌کاری نشده؟
2. راه‌حل: لایه‌ی اثبات بین SCADA و Historian
3. معماری: درج در L2.5
4. دمو: anchor یک خوانش + Merkle proof
5. air-gap روی شبکه‌ی جدا
6. نقشه: HSM در Q4 2026

### برای دادگستری (۲۰ دقیقه)

1. هدف: سیستم اثبات‌پذیر
2. جایگاه در L1-L2 (بانک) و L2.5 (صنعت)
3. آنچه اثبات می‌شود
4. آنچه ادعا نمی‌شود (FIPS, HSM, consensus)
5. زنجیره‌ی اعتماد: Authority → Switch → Auditor
6. دمو زنده
7. الزامات نصب

### سؤال‌های احتمالی

**Q: تفاوت با بلاکچین عمومی؟**
A: عمومی نیستیم — لایه‌ی اثبات درون سازمان.

**Q: تفاوت با HSM؟**
A: HSM زیرساخت (L0) است — از آن استفاده خواهیم کرد.

**Q: چه چیزی جلوی دستکاری خود Horizon را می‌گیرد؟**
A: هر بلاک با کلید Authority امضا می‌شود. کلید روی USB آفلاین است.

**Q: اگر USB گم شود؟**
A: کلید قابل rotate است، امضاهای قبلی معتبر می‌مانند. Shamir split می‌شود.

**Q: با CBS فعلی تداخل دارد؟**
A: هیچ. فقط خواندن. اتصال read-only.

---

## لینک‌های رسمی

- مخزن: https://github.com/beaconchain-horizon/horizon-core
- قرارداد رمزنگاری: https://github.com/beaconchain-horizon/horizon-core/blob/main/docs/CRYPTO-CONTRACT.md
- راهنمای دمو: https://github.com/beaconchain-horizon/horizon-core/blob/main/docs/V14-DEMO.md
- ارزیابی امنیتی: https://beaconchain-horizon.github.io/horizon-security-assessment/
- امضا USB: https://beaconchain-horizon.github.io/horizon-security-assessment/offline-license.html

---

**© 2026 Horizon — Mahdi Amoli Moghaddam**
