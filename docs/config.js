// ═══════════════════════════════════════════════════
//  HORIZON CORE — Configuration
// ═══════════════════════════════════════════════════
window.HORIZON_CONFIG = {
  // منبع داده‌ها
  DATA_SOURCE:  'local',   // 'local' = docs/data/stats.json، 'api' = SWITCH_URL
  DATA_FILE:    'data/stats.json',

  // فقط اگه DATA_SOURCE='api' باشه استفاده می‌شه
  SWITCH_URL:   'https://beaconchain-horizon.github.io',
  BACKEND_URL:  'https://beaconchain-horizon.github.io',

  // متادیتا
  SITE_NAME:    'Horizon Angel',
  SITE_TAGLINE: 'زیرساخت بلاکچین اختصاصی ایران',
  SITE_VERSION: '3.0',

  // تماس
  CONTACT_EMAIL: 'gamma.mahdii@gmail.com',
  CONTACT_PHONE: '+98-21-9100-0000'
};

console.log('⚙️ Horizon Config | Source:', window.HORIZON_CONFIG.DATA_SOURCE);
