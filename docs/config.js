// ═══════════════════════════════════════════════════════════
//  HORIZON CORE — Configuration
//  ⚠️ فقط این فایل رو بعد از خرید دامنه تغییر بده
// ═══════════════════════════════════════════════════════════

window.HORIZON_CONFIG = {

  // 🔴 جای خالی — بعد از خرید دامنه، این دو خط رو عوض کن:
  //    مثال: https://switch.yourdomain.ir
  SWITCH_URL:   'https://horizon-switch.liara.run',
  BACKEND_URL:  'https://horizon-backend.liara.run',

  // Demo mode: true = داده ماک (نمایشی)، false = API واقعی
  DEMO_MODE:    true,

  // متادیتا
  SITE_NAME:    'Horizon Angel',
  SITE_TAGLINE: 'زیرساخت بلاکچین اختصاصی ایران',
  SITE_VERSION: '3.0',

  // تماس
  CONTACT_EMAIL: 'gamma.mahdii@gmail.com',
  CONTACT_PHONE: '+98-21-9100-0000'
};

console.log('⚙️ Horizon Config loaded | Demo:', window.HORIZON_CONFIG.DEMO_MODE);
