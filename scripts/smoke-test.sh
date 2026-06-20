#!/usr/bin/env bash
# Smoke-test the running stack through the gateway: register -> login ->
# fetch profile -> update profile -> refresh. Exits non-zero on the first
# failure. Requires: a running stack (docker compose up) and curl.
#
#   ./scripts/smoke-test.sh
#
# Optional: BASE=http://localhost:8081 ./scripts/smoke-test.sh

set -euo pipefail

BASE="${BASE:-http://localhost:8080}"
EMAIL="smoke+$(date +%s)@example.com"
PASS="supersecret123"

say() { printf "\n\033[1;35m▶ %s\033[0m\n" "$1"; }
jqget() { sed -n "s/.*\"$1\":\"\([^\"]*\)\".*/\1/p"; }   # tiny extractor, no jq needed

say "Health check"
curl -fsS "$BASE/healthz" && echo

say "Register ($EMAIL)"
REG=$(curl -fsS -X POST "$BASE/api/v1/auth/register" \
  -H 'Content-Type: application/json' \
  -d "{\"email\":\"$EMAIL\",\"name\":\"Smoke Tester\",\"password\":\"$PASS\"}")
echo "$REG"
ACCESS=$(echo "$REG" | jqget accessToken)
REFRESH=$(echo "$REG" | jqget refreshToken)
[ -n "$ACCESS" ] || { echo "no access token!"; exit 1; }

say "Get profile (/users/me)"
curl -fsS "$BASE/api/v1/users/me" -H "Authorization: Bearer $ACCESS"; echo

say "Update profile (PATCH /users/me)"
curl -fsS -X PATCH "$BASE/api/v1/users/me" \
  -H "Authorization: Bearer $ACCESS" -H 'Content-Type: application/json' \
  -d '{"name":"Smoke Tester II"}'; echo

say "Refresh tokens"
curl -fsS -X POST "$BASE/api/v1/auth/refresh" \
  -H 'Content-Type: application/json' \
  -d "{\"refreshToken\":\"$REFRESH\"}"; echo

say "Login with the new account"
curl -fsS -X POST "$BASE/api/v1/auth/login" \
  -H 'Content-Type: application/json' \
  -d "{\"email\":\"$EMAIL\",\"password\":\"$PASS\"}" >/dev/null && echo "login ok"

# ---- Commerce flow: catalog -> cart -> order -> payment saga ----------------
say "List the catalog"
curl -fsS "$BASE/api/v1/products" | head -c 300; echo

say "Add a product to the cart"
curl -fsS -X POST "$BASE/api/v1/cart/items" \
  -H "Authorization: Bearer $ACCESS" -H 'Content-Type: application/json' \
  -d '{"slug":"drift-headphones","qty":2}'; echo

say "Place an order (publishes order.placed)"
ORDER=$(curl -fsS -X POST "$BASE/api/v1/orders" -H "Authorization: Bearer $ACCESS")
echo "$ORDER"

say "Wait for the payment saga, then check order status"
sleep 3
curl -fsS "$BASE/api/v1/orders" -H "Authorization: Bearer $ACCESS"; echo

say "Check this user's payments"
curl -fsS "$BASE/api/v1/payments" -H "Authorization: Bearer $ACCESS"; echo

say "Check the notification feed (welcome + order + receipt)"
curl -fsS "$BASE/api/v1/notifications" | head -c 500; echo

printf "\n\033[1;32m✓ All smoke tests passed\033[0m\n"
