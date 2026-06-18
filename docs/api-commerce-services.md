# Cart, Order, Payment & Notification — API reference

Base URL (through the gateway): `http://localhost:8080`. Same envelope as the
rest of the platform (`{ "data": … }` / `{ "error": … }`).

Cart, order and payment endpoints require a **Bearer access token** — they are
scoped to the authenticated user. The notification feed is public (demo only).

---

## Cart service

The cart belongs to the signed-in user. Items are added by **slug**; the cart
service calls the product service to fetch the authoritative name and price (it
never trusts a price from the client). Items are identified by **productId**
(the catalog UUID) for updates/removals.

| Method | Path | Body | Purpose |
|-------|------|------|---------|
| `GET` | `/api/v1/cart` | — | Get the cart |
| `POST` | `/api/v1/cart/items` | `{ "slug": "...", "qty": 1 }` | Add (or increment) an item |
| `PATCH` | `/api/v1/cart/items/{productId}` | `{ "qty": 3 }` | Set quantity (`0` removes) |
| `DELETE` | `/api/v1/cart/items/{productId}` | — | Remove an item |
| `DELETE` | `/api/v1/cart` | — | Empty the cart |

**Cart response**
```json
{
  "data": {
    "userId": "…",
    "items": [
      { "productId": "…", "slug": "drift-headphones", "name": "Drift Wireless Headphones", "priceCents": 21900, "qty": 2 }
    ],
    "count": 2,
    "totalCents": 43800
  }
}
```

```bash
curl -s -X POST http://localhost:8080/api/v1/cart/items \
  -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
  -d '{"slug":"drift-headphones","qty":2}'
```

Errors: `401` (no token), `404` (`"that product is unavailable"`).

---

## Order service

| Method | Path | Purpose |
|-------|------|---------|
| `POST` | `/api/v1/orders` | Place an order from the current cart |
| `GET` | `/api/v1/orders` | Order history (no line items) |
| `GET` | `/api/v1/orders/{id}` | One order **with** line items |

`POST /orders` takes **no body** — it reads the user's cart (server-side),
writes the order, clears the cart, and publishes `order.placed`. The order comes
back as `pending_payment`; it flips to `paid` moments later when the payment
service reports success.

**Order response**
```json
{
  "data": {
    "id": "…",
    "status": "pending_payment",
    "totalCents": 43800,
    "currency": "USD",
    "items": [
      { "productId": "…", "slug": "drift-headphones", "name": "Drift Wireless Headphones", "priceCents": 21900, "qty": 2 }
    ],
    "createdAt": "2026-06-19T10:00:00Z",
    "updatedAt": "2026-06-19T10:00:00Z"
  }
}
```

Errors: `401`; `409 empty_cart`; `404 not_found` (on GET of someone else's / unknown order).

```bash
# place an order from the current cart
curl -s -X POST http://localhost:8080/api/v1/orders -H "Authorization: Bearer $TOKEN"
# poll until status flips to "paid"
curl -s http://localhost:8080/api/v1/orders -H "Authorization: Bearer $TOKEN"
```

---

## Payment service

| Method | Path | Purpose |
|-------|------|---------|
| `GET` | `/api/v1/payments` | The user's payments |

Payments are **created by the saga, not by an HTTP call** — the payment service
subscribes to `order.placed`, charges the order (simulated, always approves a
positive amount), and records a row. This endpoint just lists them.

```json
{
  "data": [
    { "id": "…", "orderId": "…", "amountCents": 43800, "currency": "USD",
      "status": "succeeded", "provider": "mock", "createdAt": "2026-06-19T10:00:01Z" }
  ]
}
```

---

## Notification service

| Method | Path | Purpose |
|-------|------|---------|
| `GET` | `/api/v1/notifications` | Recent "sent" notifications (newest first) |

This service has **no database and no business API** — it only subscribes to
events and "sends" messages (logged + kept in an in-memory ring buffer). The
feed lets you watch the event flow during a demo. Public for convenience; you'd
lock it down or remove it in production.

```json
{
  "data": [
    { "id": "…", "type": "receipt", "channel": "email", "recipient": "user:…",
      "subject": "Payment received", "message": "We received $438.00 for order …",
      "sentAt": "2026-06-19T10:00:01Z" }
  ]
}
```

---

## The event flow (saga)

```
order-service  ──order.placed──►  product-service     (decrement stock)
                              ├─►  payment-service     (charge) ──payment.succeeded──► order-service (mark paid)
                              └─►  notification-service (order confirmation)
payment.succeeded  ───────────────────────────────────────────►  notification-service (receipt)
user.registered    ───────────────────────────────────────────►  notification-service (welcome)
```

| Subject | Producer | Consumers |
|---------|----------|-----------|
| `order.placed` | order | product, payment, notification |
| `payment.succeeded` | payment | order, notification |
| `payment.failed` | payment | order |
| `product.stock_low` | product | (available for a re-ordering service) |
| `user.registered` | user | notification |

Each consumer declares **only the fields it needs** from an event (consumer-driven
schemas), so producers can add fields without breaking anyone.
