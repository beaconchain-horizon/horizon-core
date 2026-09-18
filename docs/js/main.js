// ═══════════════════════════════════════════════════
//  Horizon Core — Main JS
// ═══════════════════════════════════════════════════

const CFG = window.HORIZON_CONFIG;

// ─── Carousel ───
let current = 0;
const slides = document.querySelectorAll('.slide');
const dots = document.getElementById('dots');

function initCarousel() {
  slides.forEach((s, i) => {
    const badge = s.dataset.badge || '';
    const title = s.dataset.title || '';
    const desc  = s.dataset.desc  || '';
    const price = s.dataset.price || '';
    const color = s.dataset.color || '#4f9eff';
    s.style.setProperty('--c', color + '30');
    s.innerHTML = `
      <div class="badge">${badge}</div>
      <h3>${title}</h3>
      <p class="desc">${desc}</p>
      <div class="price">${price}</div>
      <a href="license.html" class="buy">خرید لایسنس</a>
    `;
    const d = document.createElement('div');
    d.className = 'dot' + (i === 0 ? ' active' : '');
    d.onclick = () => goSlide(i);
    dots.appendChild(d);
  });
}

function goSlide(n) {
  if (slides.length === 0) return;
  slides[current].classList.remove('active');
  dots.children[current].classList.remove('active');
  current = (n + slides.length) % slides.length;
  slides[current].classList.add('active');
  dots.children[current].classList.add('active');
}

setInterval(() => goSlide(current + 1), 6000);

// ─── Animate numbers ───
function animate(el, target) {
  if (!el || target === undefined) return;
  const start = 0;
  const dur = 800;
  const t0 = performance.now();
  function step(now) {
    const p = Math.min((now - t0) / dur, 1);
    const e = 1 - Math.pow(1 - p, 3);
    el.textContent = Math.floor(start + (target - start) * e).toLocaleString('fa-IR');
    if (p < 1) requestAnimationFrame(step);
  }
  requestAnimationFrame(step);
}

// ─── Load Stats ───
async function loadStats() {
  let data = null;
  const statusEl = document.getElementById('net-status');

  try {
    if (CFG.DATA_SOURCE === 'api') {
      const r = await fetch(CFG.SWITCH_URL + '/api/v1/stats');
      data = await r.json();
    } else {
      const r = await fetch(CFG.DATA_FILE + '?t=' + Date.now());
      data = await r.json();
    }
  } catch (e) {
    console.warn('data unavailable:', e);
    if (statusEl) statusEl.textContent = 'شبکه قطع';
    return;
  }

  if (statusEl) statusEl.textContent = 'شبکه فعال';

  const n = data.network || {};
  animate(document.getElementById('stat-tps'), n.tps || 0);
  animate(document.getElementById('stat-blocks'), n.chain_length || 0);
  animate(document.getElementById('stat-sensors'), n.active_sensors || 0);
  animate(document.getElementById('stat-readings'), n.total_readings || 0);

  // اگه فیلدهای عددی روی صفحه با id مشخص هستن، اینجا هم می‌تونیم آپدیت کنیم
  const ver = document.getElementById('ver');
  if (ver) ver.textContent = CFG.SITE_VERSION;
}

// ─── Init ───
document.addEventListener('DOMContentLoaded', () => {
  initCarousel();
  loadStats();
  // هر ۶۰ ثانیه یه بار
  setInterval(loadStats, 60000);
});
