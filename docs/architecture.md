# Architecture

This document explains the *why* behind AURA's structure and the trade-offs made
for a learning-oriented project. Read [the README](../README.md) first for the
big picture.

---

## Principles

### 1. One service, one responsibility, one database
Each microservice owns a single business capability **and its own data store**.
The User service owns `user-postgres`; no other service may read or write that
database. If another service needs user data, it asks over the API or reacts to
a user event. This "database-per-service" rule is what allows services to:

- be developed and deployed independently,
- choose the storage that fits (Postgres for users/orders, Redis for carts),
- scale and fail in isolation.

The cost is that data is distributed: you can't `JOIN` across services. You
compose data at the API/gateway layer or keep local read-models updated via
events. That's the fundamental microservices trade-off, and feeling it is part
of the point.

### 2. An API gateway as the single front door
NGINX is the only thing exposed to the browser. It:

- routes each path prefix to the right service (`/api/v1/auth/` → user-service),
- serves the storefront for everything else,
- gives you one place for cross-cutting edge concerns (TLS, rate limiting,
  request-id injection, gzip),
- makes the whole API **same-origin**, so the browser never does CORS.

We resolve upstream hostnames at **request time** via Docker's embedded DNS
(`resolver 127.0.0.11`), using a variable in `proxy_pass`. That means NGINX
starts even if a service is still booting, and it automatically spreads load
across new replicas when you scale. See
[`gateway/nginx/conf.d/default.conf`](../gateway/nginx/conf.d/default.conf).

### 3. Synchronous REST + asynchronous events
Two communication styles, each for the right job:

- **REST (synchronous)** for request/response the caller must wait on — the
  storefront calling `POST /auth/login`.
- **Events over NATS (asynchronous)** for "this happened, react if you care" —
  the User service publishing `user.registered`. The publisher doesn't know its
  consumers, doesn't wait for them, and isn't affected if they're down. This is
  how you add a Notification service later **without touching** the User service.

Every event uses one [envelope](../pkg/events/nats.go) with an id, subject,
source service, timestamp, and a **correlation id** (the originating request id)
so an event can be traced back to the HTTP call that caused it.

The checkout flow is a **choreographed saga** — there's no central coordinator
telling everyone what to do; each service reacts to events on its own:

```
order.placed ─► product   (decrement stock)
             ├► payment   (charge) ─► payment.succeeded ─► order (mark paid)
             └► notification (confirm)
```

Three properties make this safe and decoupled:

- **Consumer-driven schemas.** Each subscriber decodes only the fields it needs
  from an event, so a producer can add fields without breaking consumers.
- **Idempotency.** NATS core is at-most-once; JetStream (and most brokers) can
  redeliver. The payment service enforces *one payment per order* (a unique
  `order_id`), so a redelivered `order.placed` never double-charges. Design
  consumers to be safe to run twice.
- **Auth propagation for sync hops.** When the order service calls the cart
  service it **forwards the customer's JWT**, so the cart service authorizes the
  request as that user — identity flows through the call chain, it isn't faked.

The trade-off is **eventual consistency**: right after checkout an order is
`pending_payment` for a beat before it becomes `paid`. The UI embraces this (it
polls), which is exactly how real systems behave.

### 4. Stateless services + stateless auth
Services keep no session state in memory, so any request can hit any replica.
Authentication is a **signed JWT**: every service validates it locally with the
shared `JWT_SECRET` — there is no per-request call back to the User service and
no shared session store. The User service is only on the critical path at
**login** and **refresh** time.

---

## The auth design

| Concern | Decision | Why |
|--------|----------|-----|
| Password storage | bcrypt, cost 12 | Slow by design; salt embedded in the hash |
| Token format | JWT (HS256) | Self-contained; any service verifies with the shared secret |
| Access token | short-lived (15m) | Limits the blast radius of a leaked token |
| Refresh token | longer (7d), typed | Lets clients stay signed in; type-tagged so it can't be used as an access token |
| User enumeration | login returns the same error for "no such email" and "wrong password", and spends comparable time | Prevents probing which emails are registered |

**Known simplification:** refresh tokens here are stateless JWTs, so they can't
be individually revoked before they expire. The production upgrade is to store
refresh tokens (e.g. in Redis) and check/rotate them on use — a localized change
in `pkg/auth` + the service layer. Likewise the frontend stores tokens in
`localStorage` (simple, but XSS-readable); the hardened version puts the refresh
token in an `httpOnly` cookie. The storefront is written against a single
[`token-store`](../web/src/lib/token-store.ts) module to make that swap easy.

---

## Request lifecycle & observability

Each service shares the same middleware chain (`pkg/middleware`):

```
request-id  →  recover  →  access-log  →  CORS  →  [auth]  →  handler
```

- **request-id** reuses the gateway's `X-Request-ID` (or generates one) and puts
  it on the context, so every log line for a request is correlated and you can
  trace a request gateway → service → event.
- **access-log** emits one structured JSON line per request (method, path,
  status, duration, request id).
- **recover** turns a handler panic into a clean `500` instead of crashing the
  process.

Health is exposed two ways, matching Kubernetes conventions:
- `/healthz` — **liveness**: is the process up?
- `/readyz` — **readiness**: can it serve traffic (Postgres + NATS reachable)?

---

## Why these technologies

| Choice | Reason |
|-------|--------|
| **Go** for services | Tiny static binaries, fast startup, great concurrency, ideal for containers |
| **chi** router | Idiomatic, std-lib-compatible `net/http`, clean middleware |
| **pgx** | The modern, high-performance Postgres driver/pool for Go |
| **Redis** | In-memory store, perfect for hot, disposable carts (with TTL) |
| **NATS** | Dead-simple, extremely fast pub/sub; JetStream available when you need durability |
| **NGINX** | Battle-tested reverse proxy; transparent, hand-written routing you can read |
| **Next.js + React + TS** | App Router, server components, first-class TS, production-ready output |
| **Tailwind** | Fast, consistent design-system styling without leaving the markup |
| **Distroless images** | No shell/package manager in the runtime → smaller and harder to exploit |

---

## Trade-offs made for the learning context

- **Single Go module** for all services + shared `pkg`. Real teams often give
  each service its own module/repo; one module keeps the template simple and
  guarantees it compiles as a unit. The directory layout still separates
  services cleanly, so splitting later is mechanical.
- **Lean core infrastructure.** Consul, ELK, and Prometheus/Grafana are
  documented but not stood up, so the stack runs on a laptop. The code already
  emits what they need (JSON logs, health probes).
- **Compose, not Kubernetes.** Compose is the right tool to *learn* the moving
  parts. The services are written to be orchestrator-agnostic (config from env,
  health probes, stateless), so moving to Kubernetes later is about manifests,
  not rewrites.

---

## Additional components

How to grow the lean core toward a production platform.

### Service discovery — Consul
Today: Docker DNS resolves `user-service` to the service's containers.
Next: run a `consul` container, have each service register itself and its health
check, and resolve other services through Consul. The gateway's request-time
resolution pattern is already discovery-friendly.

### Centralized logging — ELK or Loki
Today: every service logs structured JSON to stdout (`pkg/logger`).
Next: add `filebeat`/`promtail` to tail container logs into
Elasticsearch+Kibana (the ELK stack) or Grafana Loki. Because logs carry a
`request_id`, you can follow one request across services.

### Metrics — Prometheus + Grafana
Today: `/healthz` + `/readyz`.
Next: add `prometheus/client_golang`, expose `/metrics` (request counts,
latencies), scrape it with a `prometheus` container, and build dashboards in
`grafana`.

### CI/CD — GitHub Actions
Each service builds and tests independently, which maps onto a matrix pipeline:
`go test ./...` → `docker build` per changed service → push to a registry →
deploy. A starter workflow is a natural next addition under `.github/workflows/`.
