// ============================================================
// Horizon Brain – AI-Powered Management Panel
// ============================================================
const API_BASE = 'https://horizon-backend.liara.run';

console.log('🧠 Horizon Brain loaded.');

document.addEventListener('DOMContentLoaded', function() {
    const statusDiv = document.getElementById('aiStatus') || document.createElement('div');
    if (statusDiv) {
        checkAIStatus();
    }
});

async function checkAIStatus() {
    try {
        const resp = await fetch(`${API_BASE}/api/v1/health`, { signal: AbortSignal.timeout(3000) });
        const statusEl = document.getElementById('aiStatus');
        if (statusEl) {
            if (resp.ok) {
                statusEl.textContent = '✅ AI Connected';
                statusEl.style.color = '#4ade80';
            } else {
                statusEl.textContent = '❌ AI Offline';
                statusEl.style.color = '#f87171';
            }
        }
    } catch (e) {
        const statusEl = document.getElementById('aiStatus');
        if (statusEl) {
            statusEl.textContent = '❌ AI Offline';
            statusEl.style.color = '#f87171';
        }
    }
}

// If you need a chat function, use this:
async function askAI(question) {
    try {
        const resp = await fetch(`${API_BASE}/api/v1/ai/ask`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ question })
        });
        const data = await resp.json();
        return data.answer || data;
    } catch (e) {
        return '⚠️ AI service unavailable.';
    }
}
