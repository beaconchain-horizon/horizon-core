// ============================================================
// Horizon Core Engine - Shared API Client
// ============================================================

const API_BASE = 'https://horizon-backend.liara.run';
const SWITCH_BASE = 'https://horizon-switch.liara.run';

// ---------- Helpers ----------
async function apiGet(endpoint) {
    try {
        const r = await fetch(`${API_BASE}${endpoint}`);
        if (!r.ok) throw new Error(`HTTP ${r.status}`);
        return await r.json();
    } catch (e) { console.error(`GET ${endpoint}`, e); return null; }
}

async function apiPost(endpoint, data) {
    try {
        const r = await fetch(`${API_BASE}${endpoint}`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(data)
        });
        if (!r.ok) throw new Error(`HTTP ${r.status}`);
        return await r.json();
    } catch (e) { console.error(`POST ${endpoint}`, e); return null; }
}

async function apiSwitch(endpoint) {
    try {
        const r = await fetch(`${SWITCH_BASE}${endpoint}`);
        if (!r.ok) throw new Error(`HTTP ${r.status}`);
        return await r.json();
    } catch (e) { console.error(`SWITCH ${endpoint}`, e); return null; }
}

// ---------- Health ----------
async function checkBackendHealth() { return await apiGet('/api/v1/health'); }
async function checkSwitchHealth() { return await apiSwitch('/health'); }
async function getBenchmark() { return await apiSwitch('/benchmark'); }
async function getBlockchainStats() { return await apiSwitch('/stats'); }

// ---------- License ----------
async function generateLicense(productId, userId, volume, duration) {
    return await apiPost('/api/v1/license/generate', {
        product_id: productId, user_id: userId, volume: volume, duration: duration
    });
}
async function verifyLicense(licenseId) {
    return await apiPost('/api/v1/license/verify', { license: licenseId });
}
async function listLicenses() { return await apiGet('/api/v1/licenses'); }

// ---------- Transactions ----------
async function createTransaction(amount, type) {
    return await apiPost('/api/v1/transactions', { amount: amount, type: type, status: 'confirmed' });
}
async function listTransactions() { return await apiGet('/api/v1/transactions'); }

// ---------- Customers ----------
async function createCustomer(fullName, email, wallet) {
    return await apiPost('/api/v1/customers', { full_name: fullName, email: email, wallet: wallet });
}
async function listCustomers() { return await apiGet('/api/v1/customers'); }

// ---------- Payments ----------
async function createPayment(customerId, amount, currency) {
    return await apiPost('/api/v1/payments', { customer_id: customerId, amount: amount, currency: currency, status: 'pending' });
}
async function listPayments() { return await apiGet('/api/v1/payments'); }

// ---------- Toolbox ----------
async function encryptAES(key, plaintext) { return await apiPost('/api/v1/toolbox/encrypt', { key, plaintext }); }
async function decryptAES(key, ciphertext) { return await apiPost('/api/v1/toolbox/decrypt', { key, ciphertext }); }

// ---------- AI ----------
async function askAI(question) { return await apiPost('/api/v1/ai/ask', { question }); }

// ---------- Export ----------
window.HorizonAPI = {
    apiGet, apiPost, apiSwitch,
    checkBackendHealth, checkSwitchHealth, getBenchmark, getBlockchainStats,
    generateLicense, verifyLicense, listLicenses,
    createTransaction, listTransactions,
    createCustomer, listCustomers,
    createPayment, listPayments,
    encryptAES, decryptAES, askAI,
    API_BASE, SWITCH_BASE
};

console.log('✅ Horizon API loaded');
cd ~/Beaconchain/horizon-core && \
rm -f frontend/index.html.save && \
mkdir -p frontend/js && \
cat > frontend/js/api.js << 'EOF'
// ============================================================
// Horizon Core Engine - Shared API Client
// ============================================================

const API_BASE = 'https://horizon-backend.liara.run';
const SWITCH_BASE = 'https://horizon-switch.liara.run';

// ---------- Helpers ----------
async function apiGet(endpoint) {
    try {
        const r = await fetch(`${API_BASE}${endpoint}`);
        if (!r.ok) throw new Error(`HTTP ${r.status}`);
        return await r.json();
    } catch (e) { console.error(`GET ${endpoint}`, e); return null; }
}

async function apiPost(endpoint, data) {
    try {
        const r = await fetch(`${API_BASE}${endpoint}`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(data)
        });
        if (!r.ok) throw new Error(`HTTP ${r.status}`);
        return await r.json();
    } catch (e) { console.error(`POST ${endpoint}`, e); return null; }
}

async function apiSwitch(endpoint) {
    try {
        const r = await fetch(`${SWITCH_BASE}${endpoint}`);
        if (!r.ok) throw new Error(`HTTP ${r.status}`);
        return await r.json();
    } catch (e) { console.error(`SWITCH ${endpoint}`, e); return null; }
}

// ---------- Health ----------
async function checkBackendHealth() { return await apiGet('/api/v1/health'); }
async function checkSwitchHealth() { return await apiSwitch('/health'); }
async function getBenchmark() { return await apiSwitch('/benchmark'); }
async function getBlockchainStats() { return await apiSwitch('/stats'); }

// ---------- License ----------
async function generateLicense(productId, userId, volume, duration) {
    return await apiPost('/api/v1/license/generate', {
        product_id: productId, user_id: userId, volume: volume, duration: duration
    });
}
async function verifyLicense(licenseId) {
    return await apiPost('/api/v1/license/verify', { license: licenseId });
}
async function listLicenses() { return await apiGet('/api/v1/licenses'); }

// ---------- Transactions ----------
async function createTransaction(amount, type) {
    return await apiPost('/api/v1/transactions', { amount: amount, type: type, status: 'confirmed' });
}
async function listTransactions() { return await apiGet('/api/v1/transactions'); }

// ---------- Customers ----------
async function createCustomer(fullName, email, wallet) {
    return await apiPost('/api/v1/customers', { full_name: fullName, email: email, wallet: wallet });
}
async function listCustomers() { return await apiGet('/api/v1/customers'); }

// ---------- Payments ----------
async function createPayment(customerId, amount, currency) {
    return await apiPost('/api/v1/payments', { customer_id: customerId, amount: amount, currency: currency, status: 'pending' });
}
async function listPayments() { return await apiGet('/api/v1/payments'); }

// ---------- Toolbox ----------
async function encryptAES(key, plaintext) { return await apiPost('/api/v1/toolbox/encrypt', { key, plaintext }); }
async function decryptAES(key, ciphertext) { return await apiPost('/api/v1/toolbox/decrypt', { key, ciphertext }); }

// ---------- AI ----------
async function askAI(question) { return await apiPost('/api/v1/ai/ask', { question }); }

// ---------- Export ----------
window.HorizonAPI = {
    apiGet, apiPost, apiSwitch,
    checkBackendHealth, checkSwitchHealth, getBenchmark, getBlockchainStats,
    generateLicense, verifyLicense, listLicenses,
    createTransaction, listTransactions,
    createCustomer, listCustomers,
    createPayment, listPayments,
    encryptAES, decryptAES, askAI,
    API_BASE, SWITCH_BASE
};

console.log('✅ Horizon API loaded');
