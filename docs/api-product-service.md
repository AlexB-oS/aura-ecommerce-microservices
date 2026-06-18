# Product service — API reference

Base URL (through the gateway): `http://localhost:8080`

Same conventions as the rest of the platform: JSON bodies, success wrapped in
`{ "data": … }`, errors in `{ "error": { "code", "message", "fields?" } }`.

Reads are **public**. Writes require a **Bearer access token whose `role` claim
is `admin`** (see [Testing the admin endpoints](#testing-the-admin-endpoints)).

---

## GET /api/v1/products

List active catalog products, newest first.

**Query parameters** (all optional)

| Param | Meaning |
|-------|---------|
| `category` | exact category match, e.g. `Audio` |
| `q` | case-insensitive search over name + tagline |
| `inStock` | `true` to only return products with stock > 0 |
| `limit` | page size (default 24, max 100) |
| `offset` | rows to skip (pagination) |

**Response `200 OK`**
```json
{
  "data": [
    {
      "id": "0f9c…",
      "slug": "drift-headphones",
      "name": "Drift Wireless Headphones",
      "tagline": "40h playback, adaptive noise cancelling, zero fatigue.",
      "description": "Over-ear wireless headphones …",
      "category": "Audio",
      "priceCents": 21900,
      "currency": "USD",
      "stock": 25,
      "badge": "Best seller",
      "createdAt": "2026-06-17T10:00:00Z",
      "updatedAt": "2026-06-17T10:00:00Z"
    }
  ]
}
```

```bash
curl -s "http://localhost:8080/api/v1/products?category=Audio&inStock=true"
```

---

## GET /api/v1/products/{slug}

Fetch a single product by its slug.

**Response `200 OK`** — a single product object (same shape as above).
**Errors**: `404 not_found`.

```bash
curl -s http://localhost:8080/api/v1/products/drift-headphones
```

---

## POST /api/v1/products  🔒 admin

Create a product. The slug is generated from the name.

**Request**
```json
{
  "name": "Aurora Desk Mat",
  "tagline": "A 4mm felt mat that makes any desk feel premium.",
  "description": "Optional longer copy…",
  "category": "Desk",
  "priceCents": 4500,
  "stock": 100,
  "badge": "New"
}
```
Validation: `name` 2–120 · `category` 2–60 · `priceCents` > 0 · `stock` ≥ 0.

**Response `201 Created`** — the created product. Publishes `product.created`.

**Errors**: `401`/`403` (not an admin), `409 slug_taken`, `422 validation_failed`.

```bash
curl -s -X POST http://localhost:8080/api/v1/products \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"name":"Aurora Desk Mat","category":"Desk","priceCents":4500,"stock":100}'
```

---

## POST /api/v1/products/{id}/stock  🔒 admin

Adjust stock by a delta (positive to restock, negative to reduce). Uses the
product **id** (UUID), not the slug.

**Request**
```json
{ "delta": -1 }
```

**Response `200 OK`** — the updated product. When a reduction drops stock to ≤ 5,
a `product.stock_low` event is published.

**Errors**: `401`/`403`, `404 not_found`, `409 insufficient_stock` (would go
negative).

```bash
curl -s -X POST http://localhost:8080/api/v1/products/<ID>/stock \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"delta":-1}'
```

---

## Testing the admin endpoints

New accounts are created with the `customer` role, so registration alone can't
hit the write endpoints. Promote your user to `admin` directly in the user
database, then sign in again to get a token whose `role` claim is `admin`:

```bash
# 1. Register/login once so the user row exists, then promote it:
docker compose exec user-postgres \
  psql -U aura -d users -c "UPDATE users SET role='admin' WHERE email='you@example.com';"

# 2. Log in again to get a fresh token that carries role=admin:
ADMIN_TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"you@example.com","password":"supersecret"}' \
  | sed -n 's/.*"accessToken":"\([^"]*\)".*/\1/p')

# 3. Use it:
curl -s -X POST http://localhost:8080/api/v1/products \
  -H "Authorization: Bearer $ADMIN_TOKEN" -H 'Content-Type: application/json' \
  -d '{"name":"Test Product","category":"Misc","priceCents":1999,"stock":5}'
```

---

## Events published

| Subject | When | Payload |
|---------|------|---------|
| `product.created` | a product is added | `{ "id", "slug", "name", "priceCents" }` |
| `product.stock_low` | a reduction drops stock to ≤ 5 | `{ "id", "slug", "stock" }` |
