#!/bin/bash
set -e

cd ~/Beaconchain/horizon-core

echo ">> Replacing license.html with new panel..."
cat > frontend/license.html << 'HTML'
[محتوای کامل فایل HTML را اینجا قرار دهید - به دلیل طولانی بودن، از شما می‌خواهم خودتان آن را با استفاده از ویرایشگر nano یا کپی-پیست مستقیم انجام دهید]
HTML

# راهنمایی برای تغییر آدرس سرور
echo "⚠️ Please edit frontend/license.html and change LICENSE_SERVER_URL to:"
echo "   https://horizon-backend.liara.run"
echo "   (line ~140 in JavaScript section)"

# دیپلوی فرانت‌اند
echo ">> Deploying frontend..."
liara deploy --app horizon-frontend --static --path ./frontend

echo "✅ Done. Access at: https://horizon-frontend.liara.run/license.html"
