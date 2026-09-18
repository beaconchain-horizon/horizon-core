// ═══════════════════════════════════════════════════════════
//  Horizon Core — Main JS
//  در حالت دمو، داده‌های ماک استفاده می‌شود
// ═══════════════════════════════════════════════════════════

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
  if (!el) return;
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
  let data;
  if (CFG.DEMO_MODE) {
    data = await fetchDemoData('stats');
  } else {
    try {
      const r = await fetch(CFG.SWITCH_URL + '/stats');
      data = await r.json();
    } catch (e) {
      console.warn('API unavailable, fallback to demo');
      data = await fetchDemoData('stats');
    }
  }
  animate(document.getElementById('stat-tps'), data.tps || 0);
  animate(document.getElementById('stat-blocks'), data.chainLength || 0);
  animate(document.getElementById('stat-sensors'), 24);
  animate(document.getElementById('stat-readings'), 128493);
}

// ─── Init ───
document.addEventListener('DOMContentLoaded', () => {
  initCarousel();
  loadStats();
  setInterval(loadStats, 15000);
  const ver = document.getElementById('ver');
  if (ver) ver.textContent = CFG.SITE_VERSION;
});
