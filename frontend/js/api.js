const API_BASE = 'https://horizon-backend.liara.run';
async function apiGet(endpoint) {
    const res = await fetch(API_BASE + endpoint);
    if (!res.ok) throw new Error('Network response was not ok');
    return res.json();
}
async function apiPost(endpoint, data) {
    const res = await fetch(API_BASE + endpoint, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(data)
    });
    if (!res.ok) throw new Error('Network response was not ok');
    return res.json();
}
