# User service — API reference

Base URL (through the gateway): `http://localhost:8080`

**Conventions**
- All bodies are JSON.
- Success responses are wrapped: `{ "data": … }`.
- Error responses are wrapped: `{ "error": { "code", "message", "fields?" } }`.
- Authenticated endpoints require `Authorization: Bearer <accessToken>`.

---

## POST /api/v1/auth/register

Create a new account. Returns the user and a fresh token pair (you are signed in
immediately). Publishes a `user.registered` event.

**Request**
```json
{
  "email": "ada@example.com",
  "name": "Ada Lovelace",
  "password": "supersecret"
}
```
Validation: `email` valid & ≤254 chars · `name` 2–80 chars · `password` 8–128 chars.

**Response `201 Created`**
```json
{
  "data": {
    "user": {
      "id": "0f9c…",
      "email": "ada@example.com",
      "name": "Ada Lovelace",
      "role": "customer",
      "createdAt": "2026-06-17T10:00:00Z",
      "updatedAt": "2026-06-17T10:00:00Z"
    },
    "accessToken": "eyJhbGci…",
    "refreshToken": "eyJhbGci…",
    "tokenType": "Bearer",
    "expiresIn": 900
  }
}
```

**Errors**
- `422 validation_failed` — with a `fields` map, e.g. `{"password":"must be at least 8 characters"}`.
- `409 email_taken` — that email is already registered.

```bash
curl -s -X POST http://localhost:8080/api/v1/auth/register \
  -H 'Content-Type: application/json' \
  -d '{"email":"ada@example.com","name":"Ada Lovelace","password":"supersecret"}'
```

---

## POST /api/v1/auth/login

Authenticate with email + password.

**Request**
```json
{ "email": "ada@example.com", "password": "supersecret" }
```

**Response `200 OK`** — same shape as register (user + tokens).

**Errors**
- `401 unauthorized` — `"invalid email or password"` (returned for both a wrong
  password and an unknown email, by design).

```bash
curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"ada@example.com","password":"supersecret"}'
```

---

## POST /api/v1/auth/refresh

Exchange a valid refresh token for a new token pair (rotation).

**Request**
```json
{ "refreshToken": "eyJhbGci…" }
```

**Response `200 OK`**
```json
{
  "data": {
    "accessToken": "eyJhbGci…",
    "refreshToken": "eyJhbGci…",
    "tokenType": "Bearer",
    "expiresIn": 900
  }
}
```

**Errors**
- `401 unauthorized` — refresh token missing, expired, malformed, or an access
  token was passed instead.

---

## GET /api/v1/users/me  🔒

Return the authenticated user's profile.

**Headers**: `Authorization: Bearer <accessToken>`

**Response `200 OK`**
```json
{
  "data": {
    "id": "0f9c…",
    "email": "ada@example.com",
    "name": "Ada Lovelace",
    "role": "customer",
    "createdAt": "2026-06-17T10:00:00Z",
    "updatedAt": "2026-06-17T10:00:00Z"
  }
}
```

**Errors**: `401 unauthorized` — missing/invalid/expired token.

```bash
curl -s http://localhost:8080/api/v1/users/me \
  -H "Authorization: Bearer $TOKEN"
```

---

## PATCH /api/v1/users/me  🔒

Update mutable profile fields. Publishes a `user.updated` event.

**Request**
```json
{ "name": "Ada, Countess of Lovelace" }
```

**Response `200 OK`** — the updated user object (same shape as `GET /me`).

**Errors**: `401 unauthorized`; `422 validation_failed` (name 2–80 chars).

```bash
curl -s -X PATCH http://localhost:8080/api/v1/users/me \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"name":"Ada, Countess of Lovelace"}'
```

---

## Health endpoints

| Endpoint | Meaning | Healthy response |
|----------|---------|------------------|
| `GET /healthz` | Liveness — process is up | `200 {"data":{"status":"ok"}}` |
| `GET /readyz` | Readiness — DB + event bus reachable | `200 {"data":{"database":"ok","events":"ok"}}` |

`/readyz` returns `503` with the failing dependency marked `"unreachable"` when
Postgres or NATS is down.

---

## Events published

Consumed off NATS by other services. Envelope: see [`pkg/events`](../pkg/events/nats.go).

| Subject | When | Payload |
|---------|------|---------|
| `user.registered` | after a successful signup | `{ "id", "email", "name" }` |
| `user.updated` | after a profile change | `{ "id", "name" }` |

Each event also carries `id`, `source`, `time`, and `correlation_id` (the
originating request id) in the envelope.
