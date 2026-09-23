// ============================================================
// Horizon Angel - Offline Storage Manager
// همه چیز در مرورگر ذخیره می‌شه - دائمی
// ============================================================

window.HZ_OFFLINE = {
    // ============ localStorage: دائمی ============
    save: function(key, value) {
        try {
            localStorage.setItem('hz_' + key, JSON.stringify(value));
            return true;
        } catch(e) {
            console.error('localStorage save failed:', e);
            return false;
        }
    },
    load: function(key, defaultVal) {
        try {
            const val = localStorage.getItem('hz_' + key);
            return val ? JSON.parse(val) : defaultVal;
        } catch(e) {
            return defaultVal;
        }
    },
    remove: function(key) {
        localStorage.removeItem('hz_' + key);
    },

    // ============ IndexedDB: برای داده‌های حجیم ============
    db: null,
    initDB: function() {
        return new Promise((resolve) => {
            const req = indexedDB.open('horizon_angel', 1);
            req.onupgradeneeded = (e) => {
                const db = e.target.result;
                if (!db.objectStoreNames.contains('blocks')) {
                    db.createObjectStore('blocks', {keyPath: 'index'});
                }
                if (!db.objectStoreNames.contains('licenses')) {
                    db.createObjectStore('licenses', {keyPath: 'id'});
                }
                if (!db.objectStoreNames.contains('transactions')) {
                    db.createObjectStore('transactions', {keyPath: 'id'});
                }
            };
            req.onsuccess = (e) => {
                this.db = e.target.result;
                resolve(this.db);
            };
        });
    },
    saveBlocks: async function(blocks) {
        if (!this.db) await this.initDB();
        const tx = this.db.transaction('blocks', 'readwrite');
        const store = tx.objectStore('blocks');
        blocks.forEach(b => store.put(b));
        return new Promise(r => tx.oncomplete = r);
    },
    loadBlocks: async function() {
        if (!this.db) await this.initDB();
        const tx = this.db.transaction('blocks', 'readonly');
        const store = tx.objectStore('blocks');
        const req = store.getAll();
        return new Promise(r => req.onsuccess = () => r(req.result));
    },
    saveLicenses: async function(licenses) {
        if (!this.db) await this.initDB();
        const tx = this.db.transaction('licenses', 'readwrite');
        const store = tx.objectStore('licenses');
        licenses.forEach(l => store.put(l));
        return new Promise(r => tx.oncomplete = r);
    },
    loadLicenses: async function() {
        if (!this.db) await this.initDB();
        const tx = this.db.transaction('licenses', 'readonly');
        const store = tx.objectStore('licenses');
        const req = store.getAll();
        return new Promise(r => req.onsuccess = () => r(req.result));
    },

    // ============ Session: ورود دائمی ============
    saveSession: function(session) {
        session.saved_at = Date.now();
        session.expires_at = Date.now() + (365 * 24 * 60 * 60 * 1000); // 1 year
        this.save('session', session);
    },
    loadSession: function() {
        const s = this.load('session', null);
        if (!s) return null;
        if (s.expires_at && Date.now() > s.expires_at) {
            this.remove('session');
            return null;
        }
        return s;
    },

    // ============ Connection Status ============
    isOnline: function() {
        return navigator.onLine;
    },
    watchConnection: function(onChange) {
        window.addEventListener('online', () => onChange(true));
        window.addEventListener('offline', () => onChange(false));
    }
};

// ============ Auto-register Service Worker ============
if ('serviceWorker' in navigator) {
    window.addEventListener('load', () => {
        navigator.serviceWorker.register('/sw.js')
            .then(reg => console.log('✅ SW registered'))
            .catch(e => console.log('SW failed:', e));
    });
}

// ============ Auto-init IndexedDB ============
window.HZ_OFFLINE.initDB();

console.log('✅ Offline storage ready');
