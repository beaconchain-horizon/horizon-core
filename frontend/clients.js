// ============================================================
// Horizon — Clients Registry
// هر مشتری: اسم، Chain ID اختصاصی، نوع، Air-Gap
// ============================================================

const HORIZON_CLIENTS = {
    // ── بانک‌ها ──
    'bank_markazi': {
        name: 'بانک مرکزی ایران',
        chain_id: 'cbi-01',
        type: 'bank',
        air_gap: false,
        icon: '🏦'
    },
    'bank_melli': {
        name: 'بانک ملی ایران',
        chain_id: 'bank-melli-01',
        type: 'bank',
        air_gap: false,
        icon: '🏦'
    },
    'bank_saderat': {
        name: 'بانک صادرات ایران',
        chain_id: 'bank-saderat-01',
        type: 'bank',
        air_gap: false,
        icon: '🏦'
    },
    'bank_mellat': {
        name: 'بانک ملت',
        chain_id: 'bank-mellat-01',
        type: 'bank',
        air_gap: false,
        icon: '🏦'
    },

    // ── صنایع ──
    'niordc': {
        name: 'شرکت ملی پالایش و پخش',
        chain_id: 'niordc-01',
        type: 'industry',
        air_gap: true,
        icon: '🛢️'
    },
    'igmc': {
        name: 'مدیریت شبکه برق',
        chain_id: 'igmc-01',
        type: 'industry',
        air_gap: true,
        icon: '⚡'
    },
    'nigc': {
        name: 'شرکت ملی گاز',
        chain_id: 'nigc-01',
        type: 'industry',
        air_gap: true,
        icon: '🔥'
    },
    'nicico': {
        name: 'صنایع مس ایران',
        chain_id: 'nicico-01',
        type: 'industry',
        air_gap: true,
        icon: '⛏️'
    },
    'msc': {
        name: 'فولاد مبارکه',
        chain_id: 'msc-01',
        type: 'industry',
        air_gap: true,
        icon: '🏭'
    }
};

// ── Helpers ──
function getClient(id) {
    return HORIZON_CLIENTS[id] || {
        name: id,
        chain_id: 'unknown',
        type: 'unknown',
        air_gap: false,
        icon: '❓'
    };
}

function getClientName(id) { return getClient(id).name; }
function getClientChain(id) { return getClient(id).chain_id; }
function getClientAirGap(id) { return getClient(id).air_gap; }
function getClientType(id) { return getClient(id).type; }
