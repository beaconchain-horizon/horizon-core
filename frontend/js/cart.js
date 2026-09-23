function getCart() {
    return JSON.parse(localStorage.getItem('horizon_cart') || '[]');
}
function saveCart(cart) {
    localStorage.setItem('horizon_cart', JSON.stringify(cart));
    updateCartBadge();
}
function addToCart(productId) {
    const cart = getCart();
    const existing = cart.find(item => item.id === productId);
    if (existing) {
        existing.quantity += 1;
    } else {
        cart.push({ id: productId, quantity: 1 });
    }
    saveCart(cart);
    alert('محصول به سبد خرید اضافه شد');
}
function removeFromCart(productId) {
    let cart = getCart();
    cart = cart.filter(item => item.id !== productId);
    saveCart(cart);
}
function updateCartBadge() {
    const cart = getCart();
    const total = cart.reduce((sum, item) => sum + item.quantity, 0);
    const badge = document.getElementById('cart-count');
    if (badge) badge.textContent = total;
}
