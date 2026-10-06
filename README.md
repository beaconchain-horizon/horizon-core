# Horizon Core — Industrial Platform

**Offline-First Tamper-Evident Ledger for Banking, Industry, and Air-Gapped Environments**

نسخه: 3.0 | تاریخ: ۲۰۲۶-۱۰-۰۶ | کامیت: 4ba78ac

---

## What It Is — معرفی

Horizon Core یک پلتفرم بلاکچین سازمانی است که همزمان سه حوزه را پوشش می‌دهد:

| # | حوزه | کاربرد |
|---|------|--------|
| ۱ | **بانکداری** | تراکنش بین‌بانکی، تسویه، Dual Control |
| ۲ | **صنعت** | Anchor داده سنسور، Merkle Proof، SCADA |
| ۳ | **Air-Gap** | جداسازی کامل برای سایت‌های حیاتی |

سیستم بر پایه‌ی یک **Ledger تغییرناپذیر** بنا شده که با **رمزنگاری ECDSA P-256** امضا می‌شود و در برابر دستکاری مقاوم است.

---

## Core Capabilities — قابلیت‌های اصلی

### Ledger & Blockchain

- ✅ **Single-node append-only ledger** — دفتر کل فقط-افزودنی
- ✅ **Merkle-root hash chain** — زنجیر هش Merkle (دودویی + سه‌دویی)
- ✅ **RFC 6962 compliant** — استاندارد بین‌المللی Merkle
- ✅ **ECDSA P-256 block signatures** — امضای هر بلاک (FIPS 186-5)
- ✅ **Immutable audit log** — لاگ تغییرناپذیر
- ✅ **int64 Money storage** — دقت بانکی (بدون خطای IEEE 754)

### Industrial Monitoring — پایش صنعتی

- ✅ **Sensor data anchoring** — anchor داده سنسورها روی زنجیر
- ✅ **30-second anchor cycle** — هر ۳۰ ثانیه (bounded to 50K readings)
- ✅ **Merkle Proof for each reading** — اثبات اصالت هر داده
- ✅ **Tamper detection** — تشخیص هر تغییر غیرمجاز
- ✅ **SCADA / Historian integration ready** — آماده ادغام
- ✅ **Batch processing** — پردازش دسته‌ای تا ۵۰,۰۰۰ خوانش

### Air-Gap Mode — حالت جداسازی

- ✅ **Outbound HTTP block** — قطع کامل ارتباطات خروجی
- ✅ **Shamir Secret Sharing** — تقسیم کلید بین چند طرف (GF(2^8))
- ✅ **Offline operation** — کارکرد بدون اینترنت
- ✅ **Controlled transfer** — انتقال کنترل‌شده داده
- ✅ **دیود نرم‌افزاری** — software diode (77 خط کد)

### Security — امنیت

- ✅ **Session Token** — تک‌مصرفه، ۵ دقیقه، ضد replay
- ✅ **Dual Control** — تأیید دوگانه برای مبالغ بالای ۱M ریال
- ✅ **Sanitize Middleware** — مسدودسازی ۳۰ الگوی خطرناک
- ✅ **Rate Limiter** — ۵۰۰,۰۰۰ req/s
- ✅ **Key Rotation** — چرخش کلید
- ✅ **Grace Period** — دوره‌ی گذشت برای لایسنس
- ✅ **License Enforcement** — اجبار لایسنس سالانه
- ✅ **192 penetration tests** — امتیاز ۹۸٪

### Performance — کارایی

- ⚡ **TPS: 10,000+** — روی Windows 8-core (dev machine)
- ⚡ **Latency: < 1 second**
- ⚡ **Zero data loss** — صفر از دست رفتن داده
- ⚡ **Success rate: 100%** — درصد موفقیت کامل

---

## Industrial Use Cases — کاربردهای صنعتی

### پالایشگاه (سرچشمه، آبادان، بندرعباس)

| قابلیت | کاربرد |
|---|---|
| Anchor خوانش سنسور | ثبت فشار، دما، جریان |
| Merkle Proof | اثبات اصالت داده به رگولاتور |
| Air-Gap | جداسازی کامل از شبکه IT |
| Audit Log | حسابرسی داخلی و خارجی |

### نیروگاه (برق حرارتی، هسته‌ای، تجدیدپذیر)

| قابلیت | کاربرد |
|---|---|
| ثبت لحظه‌ای پارامترها | ولتاژ، جریان، فرکانس |
| تشخیص تغییر | هرگونه دستکاری در داده |
| Immutable Ledger | گزارش رسمی برای وزارت نیرو |
| Shamir Secret | حفاظت از کلید اصلی |

### خطوط نفت و گاز

| قابلیت | کاربرد |
|---|---|
| Flow meter anchoring | ثبت جریان در هر ۳۰s |
| Batch Proof | اثبات کل خط لوله |
| Offline Mode | کار در مناطق دورافتاده |
| SCADA Integration | ادغام با سیستم موجود |

### صنایع حیاتی (هسته‌ای، دفاعی، هوافضا)

| قابلیت | کاربرد |
|---|---|
| Full Air-Gap | بدون هیچ اتصال خارجی |
| Hardware-bound License | لایسنس مقید به سخت‌افزار |
| Dual Control | تأیید چندنفره |
| Key Escrow | نگهداری امن کلید |

---

## What It Is NOT — محدودیت‌های صریح

**صداقت فنی، اعتماد می‌آورد:**

- ❌ Not distributed consensus — تک‌گره، نه اجماع توزیع‌شده (در Roadmap)
- ❌ Not a public blockchain — خصوصی، نه عمومی (مثل Bitcoin/Ethereum)
- ❌ Not a replacement for Fabric/Corda — مکمل است، نه جایگزین
- ❌ Not HSM-based (yet) — نرم‌افزاری، نه سخت‌افزار (خرید از بانک)
- ❌ Not FIPS certified (yet) — استفاده می‌کند، ولی گواهی رسمی ندارد

---

## Technical Stack — پشته‌ی فنی

| لایه | تکنولوژی |
|---|---|
| زبان | Go 1.24 |
| HTTP | Gin |
| Database | SQLite (WAL mode) + PostgreSQL |
| Merkle | RFC 6962 |
| Crypto | ECDSA P-256, Shamir, SHA-256 |
| Container | Docker |
| Deployment | Liara, Linux |

---

## Roadmap — نقشه راه

| فاز | قابلیت | وضعیت |
|---|---|---|
| ۱ | Single-node Ledger | ✅ تکمیل |
| ۲ | Air-Gap Mode | ✅ تکمیل |
| ۳ | Dual Control | ✅ تکمیل |
| ۴ | Industrial Anchor | ✅ تکمیل |
| ۵ | Multi-node Consensus | ⏳ Q1 2027 |
| ۶ | HSM Integration | ⏳ Q4 2026 |
| ۷ | ISO 27001 | ⏳ Q2 2027 |
| ۸ | FIPS Certification | ⏳ Q3 2027 |

---

## Docs — مستندات

- `docs/ARCHITECTURE.md` — معماری سیستم
- `docs/CERTIFICATE.md` — گواهی‌ها
- `docs/RFC-6962-COMPLIANCE.md` — انطباق با RFC 6962
- `docs/review/CONSENSUS.md` — مکانیزم اجماع
- `docs/review/AIRGAP.md` — Air-Gap Mode
- `docs/review/BENCHMARK-SPEC.md` — بنچمارک
- `docs/review/SECURITY-ROADMAP.md` — نقشه امنیتی
- `docs/review/REVIEW-STATUS.md` — وضعیت بازبینی
- `threat-model/THREAT-MODEL.md` — مدل تهدید STRIDE
- `sbom/SBOM.md` — فهرست کتابخانه‌ها

---

## Penetration Testing — تست‌های نفوذ

| دسته | تعداد | نتیجه |
|---|---:|---|
| SQL Injection | ۳۰ | ✅ مسدود |
| XSS | ۲۵ | ✅ مسدود |
| Command Injection | ۲۰ | ✅ مسدود |
| Path Traversal | ۱۵ | ✅ مسدود |
| Auth Bypass | ۲۰ | ✅ مسدود |
| JWT Attacks | ۱۰ | ✅ مسدود |
| Header Injection | ۱۵ | ✅ مسدود |
| HTTP Methods | ۱۲ | ✅ مسدود |
| Oversize Payload | ۱۵ | ✅ مسدود |
| SSRF | ۱۰ | ✅ مسدود |
| XXE | ۵ | ✅ مسدود |
| Business Logic | ۱۵ | ✅ مسدود |
| **جمع** | **۱۹۲** | **امتیاز ۹۸٪** |

---

## License — مجوز

**Commercial — All Rights Reserved**

- مجوز سالانه (یک‌ساله)
- Enforcer خودکار (قطع پس از انقضا)
- Bound to Hardware ID
- See `COMMERCIAL_LICENSE.md`

---

## Contact — تماس

**© 2026 Horizon — Mahdi Amoli Moghaddam**

- 📦 مخزن اصلی: `github.com/beaconchain-horizon/horizon-core` (PRIVATE)
- 🔐 مخزن امنیتی: `github.com/beaconchain-horizon/horizon-security-assessment`
