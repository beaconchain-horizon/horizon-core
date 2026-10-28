#!/usr/bin/env bash
BE="http://localhost:8080"
TOK="12345678901234567890123456789012"

echo "Restoring Blocks 1 to 4..."
sleep 2

echo "-> Sending GOLD-001..."
curl -s -X POST "$BE/api/v1/license/register" -H "Content-Type: application/json" -H "X-Admin-Key: $TOK" -d '{"license_id":"HORIZON-GOLD-001","user_id":"bank-markazi","product_id":"horizon-core","volume":50000000,"signature":"110b3a136ab4c8ebe598831ffba8a1f37adb814a5ea67c52a0c1e7b01aa75399130e99b0a2af61c7c8ec99a1256049c6f7ae9fd0b0754412f54c06d6e0a434f8","signed_data":"{\"license_id\":\"HORIZON-GOLD-001\",\"user_id\":\"bank-markazi\",\"product_id\":\"horizon-core\",\"volume\":50000000,\"duration\":365,\"merkle_root\":\"7679585b7f2a1007ecacca0bd801641e37cab96453d774d7c21817b4dfda638d\",\"issued_at\":1792152178,\"expires_at\":1823688178,\"hardware_id\":\"\"}","public_key":"04cf6aaaba8e7a6fddde63a2e2588553dc95503d2de6b2a5c78f4a30a0b91b0aca4d28d06fc1ccbaf730abe1dcda7c29726dab18008353b521c9253c04b1641214"}'
echo ""
sleep 3

echo "-> Sending GOLD-002..."
curl -s -X POST "$BE/api/v1/license/register" -H "Content-Type: application/json" -H "X-Admin-Key: $TOK" -d '{"license_id":"HORIZON-GOLD-002","user_id":"bank-markazi","product_id":"horizon-core","volume":50000000,"signature":"fd7ec3ff42d1c5db1fc14d56408b6e3cfbd62a330a7eec0f6f85382b7adad20948641870ddd85c24ba9053040554906faa92725fe20a6b98509d947bae9736f5","signed_data":"{\"license_id\":\"HORIZON-GOLD-002\",\"user_id\":\"bank-markazi\",\"product_id\":\"horizon-core\",\"volume\":50000000,\"duration\":365,\"merkle_root\":\"ea97e0a08f11f389302bbb621b710d062dcb564d3665b72e1fde84447bd1ba25\",\"issued_at\":1792159778,\"expires_at\":1823695778,\"hardware_id\":\"\"}","public_key":"04cf6aaaba8e7a6fddde63a2e2588553dc95503d2de6b2a5c78f4a30a0b91b0aca4d28d06fc1ccbaf730abe1dcda7c29726dab18008353b521c9253c04b1641214"}'
echo ""
sleep 3

echo "-> Sending GOLD-003..."
curl -s -X POST "$BE/api/v1/license/register" -H "Content-Type: application/json" -H "X-Admin-Key: $TOK" -d '{"license_id":"HORIZON-GOLD-003","user_id":"bank-markazi","product_id":"horizon-core","volume":50000000,"signature":"5f8af4d33b0e6957bd9f131fe6d6595edd0c416be8e8a5e3a8a3cd2bda3d31ef624235a0f2c3457bb9febc8c365ad1bc99ab6ff36fee618008b0594c3678211e","signed_data":"{\"license_id\":\"HORIZON-GOLD-003\",\"user_id\":\"bank-markazi\",\"product_id\":\"horizon-core\",\"volume\":50000000,\"duration\":365,\"merkle_root\":\"d6372d10b57a40c70a4808eac836cedebd8f4028c701d3e9e49d0443b1ab30c7\",\"issued_at\":1792148281,\"expires_at\":1823684281,\"hardware_id\":\"\"}","public_key":"04cf6aaaba8e7a6fddde63a2e2588553dc95503d2de6b2a5c78f4a30a0b91b0aca4d28d06fc1ccbaf730abe1dcda7c29726dab18008353b521c9253c04b1641214"}'
echo ""
sleep 3

echo "-> Sending GOLD-004..."
curl -s -X POST "$BE/api/v1/license/register" -H "Content-Type: application/json" -H "X-Admin-Key: $TOK" -d '{"license_id":"HORIZON-GOLD-004","user_id":"bank-markazi","product_id":"horizon-core","volume":50000000,"signature":"1c70ad2e7293e37ab06728938d29db9c76c658cfc659c360805b5262316442276c80049c3d951347162909cdf64da5f37083835c1792fc5552c972b060305393","signed_data":"{\"license_id\":\"HORIZON-GOLD-004\",\"user_id\":\"bank-markazi\",\"product_id\":\"horizon-core\",\"volume\":50000000,\"duration\":365,\"merkle_root\":\"292d2e3f85f281465998355924778226f4aec9f182b1a2c3ade823cc07970c18\",\"issued_at\":1792164282,\"expires_at\":1823700282,\"hardware_id\":\"\"}","public_key":"04cf6aaaba8e7a6fddde63a2e2588553dc95503d2de6b2a5c78f4a30a0b91b0aca4d28d06fc1ccbaf730abe1dcda7c29726dab18008353b521c9253c04b1641214"}'
echo ""

echo "Restoration complete."
