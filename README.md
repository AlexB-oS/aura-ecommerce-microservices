# AURA — a scalable e-commerce platform (microservices + Docker)

[![CI](https://github.com/AlexB-oS/aura-ecommerce-microservices/actions/workflows/ci.yml/badge.svg)](https://github.com/AlexB-oS/aura-ecommerce-microservices/actions/workflows/ci.yml)
![Go](https://img.shields.io/badge/Go-1.23-00ADD8?logo=go&logoColor=white)
![Next.js](https://img.shields.io/badge/Next.js-14-000000?logo=next.js)
![Docker](https://img.shields.io/badge/Docker-Compose-2496ED?logo=docker&logoColor=white)
![License](https://img.shields.io/badge/License-MIT-blue.svg)

AURA is a **learn-by-building** reference implementation of a modern, distributed
e-commerce platform. It is intentionally built one service at a time so you can
*see* how a real microservices system fits together — gateway, services,
databases, an event bus, and a polished storefront — instead of being handed a
finished monolith.

> **Stack:** Go (services) · TypeScript + React + Next.js (storefront) ·
> NGINX (API gateway) · PostgreSQL · Redis · NATS (events) · Docker & Docker Compose.

This repository is the **foundation + a complete first service**. The User
service is implemented end-to-end as the *template* you copy to build the rest.
Everything else (Product, Cart, Order, Payment, Notification) is scaffolded into
the architecture, the gateway, and the storefront so the path forward is obvious.

---

## Table of contents

1. [What's built right now](#whats-built-right-now)
2. [Architecture at a glance](#architecture-at-a-glance)
3. [Repository layout](#repository-layout)
4. [Prerequisites](#prerequisites)
5. [Quick start](#quick-start)
6. [How a request flows](#how-a-request-flows)
7. [The User service API](#the-user-service-api)
8. [Running the tests](#running-the-tests)
9. [Scaling a service](#scaling-a-service)
10. [The roadmap — building the other services](#the-roadmap--building-the-other-services)
11. [Additional platform components (next steps)](#additional-platform-components-next-steps)
12. [What you'll learn](#what-youll-learn)
13. [Troubleshooting](#troubleshooting)

---

## What's built right now

| Area | Status | Notes |
|------|--------|-------|
| **API Gateway** (NGINX) | ✅ Live | Single entrypoint on `:8080`, request-time DNS, request-id propagation |
| **User service** (Go) | ✅ Live | Register, login, JWT refresh + roles, profile, events, tests |
| **Product service** (Go) | ✅ Live | Catalog (public) + admin RBAC writes, seeded data, decrements stock on `order.placed` |
| **Cart service** (Go) | ✅ Live | Redis-backed user carts; calls the product service to snapshot price/name |
| **Order service** (Go) | ✅ Live | Checkout from cart, order history, listens for payment results |
| **Payment service** (Go) | ✅ Live | Simulated charge on `order.placed`, idempotent, emits `payment.succeeded` |
| **Notification service** (Go) | ✅ Live | Pure event subscriber — "sends" on user/order/payment events |
| **PostgreSQL** | ✅ Live | Database-per-service (user, product, order, payment) |
| **Redis** | ✅ Live | Backs the cart service |
| **NATS** (event bus) | ✅ Live | The async backbone: order → stock/payment/notification saga |
| **Storefront** (Next.js) | ✅ Live | "AURA" design, real auth, live catalog, **working checkout** + order history |
| Service discovery (Consul) | 📋 Documented | See [next steps](#additional-platform-components-next-steps) |
| Centralized logging (ELK) | 📋 Documented | JSON logs already emitted, ready to ship |
| Metrics (Prometheus/Grafana) | 📋 Documented | Health/readiness probes in place |

---

## Architecture at a glance

```
   Browser ──►  NGINX API Gateway  :8080  (the only public port; routes by path prefix)
                        │
   ┌────────────┬───────┼─────────┬──────────┬──────────┬───────────────┐
   ▼            ▼       ▼         ▼          ▼          ▼               ▼
  web        user    product    cart       order     payment      notification
(Next.js)  service  service   service    service    service        service
   │           │        │         │          │          │               │
   │           ▼        ▼         ▼          ▼          ▼          (no database)
   │     user-pg   product-pg   redis    order-pg   payment-pg
   │
   │     ── synchronous REST ──►  cart→product (price snapshot),  order→cart (read/clear)
   │
   └───────────────────  all services publish/subscribe on  NATS  ───────────────────┐
                                                                                      │
   order.placed ──►  product (decrement stock)                                        │
                ├─►  payment (charge) ──► payment.succeeded ──► order (mark paid)      │
                └─►  notification (confirm)        │                                   │
   user.registered, payment.succeeded ────────────┴─►  notification (welcome/receipt) ┘
```

**Four ideas hold it together:**

1. **One gateway, many services.** The browser only ever talks to NGINX. It
   routes each URL prefix to the right service. Adding a service = adding a
   `location` block. Because everything is one origin, the frontend has **no
   CORS** to fight in the browser.
2. **Database-per-service.** Each service owns its data (the User service owns
   `user-postgres`). Services never reach into each other's database — only over
   the API or the event bus. This is what lets them evolve and scale apart.
3. **Event-driven decoupling.** When the User service registers someone it
   publishes a `user.registered` event to NATS. It does **not** know or care who
   listens. The Notification service (built later) just subscribes. No code
   change to the User service is needed to add a consumer.
4. **Stateless services + shared JWT secret.** Auth is a signed JWT. Every
   service can validate it locally with the shared secret — no call back to the
   User service on each request, and any service can run N replicas behind the
   gateway.

A deeper discussion of the trade-offs lives in [docs/architecture.md](docs/architecture.md).

---

## Repository layout

```
scalable_eccomerce/
├── docker-compose.yml        # orchestrates the whole stack (lean core)
├── Makefile                  # `make up`, `make test`, `make scale-user`, …
├── .env.example              # copy to .env (JWT secret, DB password, …)
├── go.mod                    # single Go module for all services + shared pkg
│
├── gateway/nginx/            # the API gateway
│   ├── nginx.conf            #   http-level config (gzip, ws upgrade map, logs)
│   └── conf.d/default.conf   #   the routing table (one block per service)
│
├── pkg/                      # shared Go libraries (used by EVERY service)
│   ├── logger/               #   structured JSON logging (slog)
│   ├── httpx/                #   JSON envelopes + typed API errors
│   ├── auth/                 #   JWT issue/verify + roles + bcrypt hashing
│   ├── middleware/           #   request-id, recover, logging, CORS, auth, RBAC
│   ├── events/               #   NATS publish/subscribe wrapper + event envelope
│   ├── restclient/           #   typed client for sync service-to-service calls
│   ├── db/                   #   Postgres pool + embedded migration runner (shared)
│   ├── config/               #   env-var helpers
│   └── validate/             #   struct-tag validation → field error map
│
├── services/                 # six independent microservices, same internal shape
│   ├── user/                 # ★ the reference service (auth, profiles)
│   ├── product/              #   catalog; subscribes to order.placed → stock
│   ├── cart/                 #   Redis-backed carts; REST-calls product service
│   ├── order/                #   checkout orchestration; saga coordinator
│   ├── payment/              #   simulated charge; subscribes to order.placed
│   └── notification/         #   pure event subscriber (no database, no API)
│       ├── cmd/server/       #   main.go: wiring + graceful shutdown + healthcheck
│       ├── internal/
│       │   ├── config/       #   env → typed config
│       │   ├── domain/       #   entities + domain errors (no infra knowledge)
│       │   ├── repository/   #   Postgres adapter + embedded migration runner
│       │   ├── service/      #   business logic (unit-tested, no DB needed)
│       │   ├── transport/http#   handlers, DTOs, router
│       │   └── events/       #   typed event payloads (the integration contract)
│       ├── migrations/       #   embedded *.sql, applied once on startup
│       └── Dockerfile        #   multi-stage → distroless, self-healthchecking
│
├── web/                      # the AURA storefront (Next.js App Router + Tailwind)
│   ├── src/app/              #   pages: /, /products, /login, /register, /account
│   ├── src/components/       #   design system (Header, ProductCard, CartDrawer…)
│   ├── src/lib/              #   API client, auth + cart contexts, types
│   └── Dockerfile            #   standalone production build
│
└── docs/                     # the deep-dive guides
    ├── architecture.md
    ├── getting-started.md
    ├── adding-a-service.md   #   ← how to build the NEXT service from this template
    └── api-user-service.md
```

The User service follows a **clean / hexagonal** layering — `domain` →
`service` → `repository`/`transport`. Dependencies point inward: the business
logic depends on interfaces, not on Postgres or HTTP. That is exactly what makes
the service unit-testable with zero infrastructure (see
[`service_test.go`](services/user/internal/service/service_test.go)).

---

## Prerequisites

You need **Docker Desktop** to run the stack, and (optionally) the Go and Node
toolchains if you want to run services or the frontend outside Docker.

- **Docker Desktop** (includes Docker Compose v2) — required.
- **Go 1.23+** — optional, for `make test` / local builds.
- **Node 20+** — optional, for `npm run dev` hot-reload on the storefront.

> On this machine none of these are installed yet — install Docker Desktop first.
> Full Windows setup steps are in [docs/getting-started.md](docs/getting-started.md).

---

## Quick start

```bash
# 1. Configure the environment (sets JWT secret, DB password, …)
cp .env.example .env
#   then edit .env and set a strong JWT_SECRET

# 2. Build & start everything
docker compose up --build          # or: make up

# 3. Open the storefront
#    → http://localhost:8080
```

That single command builds all six Go services and the Next.js storefront, and
starts NGINX, four Postgres databases, Redis and NATS. Each service runs its own
migrations automatically on startup. (First build pulls images and compiles Go —
give it a few minutes.)

**Smoke-test the API directly through the gateway:**

```bash
# Register a user
curl -s -X POST http://localhost:8080/api/v1/auth/register \
  -H 'Content-Type: application/json' \
  -d '{"email":"ada@example.com","name":"Ada Lovelace","password":"supersecret"}'

# Log in (grab the accessToken from the response)
curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"ada@example.com","password":"supersecret"}'

# Fetch your profile (replace <TOKEN>)
curl -s http://localhost:8080/api/v1/users/me \
  -H 'Authorization: Bearer <TOKEN>'
```

Stop the stack with `docker compose down` (add `-v` to also wipe the database
volume).

---

## How a request flows

**Signing up, end to end:**

1. The browser submits the register form on the Next.js storefront.
2. The request `POST /api/v1/auth/register` hits **NGINX**, which matches the
   `/api/v1/auth/` prefix and proxies to **user-service**, attaching an
   `X-Request-ID`.
3. The user service's middleware chain runs: request-id → panic-recover →
   access-log → CORS. The handler **decodes + validates** the body (bad input
   comes back as a per-field `422`).
4. The **service layer** hashes the password (bcrypt), inserts the user via the
   **repository** (a duplicate email becomes a clean `409 email_taken`), and
   **publishes** a `user.registered` event to **NATS**.
5. It issues a JWT **access + refresh** token pair and returns `201` with the
   user and tokens.
6. The storefront stores the tokens and redirects to `/account`. From now on the
   API client attaches the access token and **auto-refreshes** it on a `401`.

## The checkout saga (sync + async working together)

Placing an order is the showpiece of the architecture — it combines synchronous
calls and an asynchronous, event-driven **saga** across four services:

1. The storefront `POST /api/v1/orders` reaches **order-service** (authenticated).
2. **Synchronously**, the order service calls **cart-service** — *forwarding the
   customer's own JWT* — to read their cart, then writes an immutable order
   (status `pending_payment`) to its database and clears the cart.
3. It **publishes `order.placed`** to NATS and returns `201` immediately. The
   customer doesn't wait for payment.
4. **Asynchronously**, three services react to `order.placed`, none of them
   called directly by the order service:
   - **product-service** decrements stock for each line item (and may emit
     `product.stock_low`).
   - **payment-service** charges the order (simulated), idempotently (one payment
     per order, so a redelivered event never double-charges), and publishes
     **`payment.succeeded`**.
   - **notification-service** "sends" an order confirmation.
5. **order-service** subscribes to `payment.succeeded` and flips the order to
   `paid`. **notification-service** sends a receipt.

The orders page polls a couple of times so you can watch an order go
`pending_payment → paid`, and the live notification feed is at
`GET /api/v1/notifications`. Adding the notification service required **zero
changes** to the services that emit the events — that's the payoff of
event-driven decoupling.

The request id set at the gateway appears in every log line across the hop, so
you can `grep` one request through the whole system.

---

## The User service API

Base URL through the gateway: `http://localhost:8080`

| Method | Path | Auth | Purpose |
|-------|------|------|---------|
| `POST` | `/api/v1/auth/register` | — | Create an account, returns user + tokens |
| `POST` | `/api/v1/auth/login` | — | Authenticate, returns user + tokens |
| `POST` | `/api/v1/auth/refresh` | — | Exchange a refresh token for new tokens |
| `GET` | `/api/v1/users/me` | Bearer | Current user's profile |
| `PATCH` | `/api/v1/users/me` | Bearer | Update profile (name) |
| `GET` | `/healthz` | — | Liveness (process up) |
| `GET` | `/readyz` | — | Readiness (DB + event bus reachable) |

Every response is enveloped: success is `{ "data": … }`, errors are
`{ "error": { "code", "message", "fields?" } }`. Full request/response examples
are in [docs/api-user-service.md](docs/api-user-service.md).

## The Product service API

| Method | Path | Auth | Purpose |
|-------|------|------|---------|
| `GET` | `/api/v1/products` | — | List the catalog (filters: `category`, `q`, `inStock`, `limit`, `offset`) |
| `GET` | `/api/v1/products/{slug}` | — | One product by slug |
| `POST` | `/api/v1/products` | Bearer **admin** | Create a product |
| `POST` | `/api/v1/products/{id}/stock` | Bearer **admin** | Adjust stock (`{ "delta": -1 }`) |

Reads are public; writes require an **admin** JWT (the `role` claim is checked by
`middleware.RequireRole`). The catalog comes pre-seeded with demo products. Full
examples are in [docs/api-product-service.md](docs/api-product-service.md).

## The Cart / Order / Payment / Notification APIs

All cart and order endpoints require a Bearer access token (they're user-scoped).

| Method | Path | Purpose |
|-------|------|---------|
| `GET` | `/api/v1/cart` | Current user's cart |
| `POST` | `/api/v1/cart/items` | Add an item (`{ "slug": "...", "qty": 1 }`) |
| `PATCH` | `/api/v1/cart/items/{productId}` | Set quantity (`0` removes) |
| `DELETE` | `/api/v1/cart/items/{productId}` | Remove an item |
| `DELETE` | `/api/v1/cart` | Empty the cart |
| `POST` | `/api/v1/orders` | Place an order from the cart |
| `GET` | `/api/v1/orders` | Order history |
| `GET` | `/api/v1/orders/{id}` | One order with line items |
| `GET` | `/api/v1/payments` | The user's payments |
| `GET` | `/api/v1/notifications` | Recent notifications (demo feed, public) |

Full examples and the event payloads are in
[docs/api-commerce-services.md](docs/api-commerce-services.md).

---

## Running the tests

The business logic is tested without a database by swapping the Postgres
repository for an in-memory fake (the payoff of depending on interfaces):

```bash
go test ./...          # or: make test
```

Dependencies are fully pinned — `go.sum` and `web/package-lock.json` are
committed, so `go test`, `docker compose build`, and CI all resolve the exact
same versions every time.

Each service's `service` package has unit tests run against in-memory fakes —
registration & token refresh (user), price validation & low-stock events
(product), cart merge & price snapshots (cart), checkout totals & cart-clearing
(order), and **idempotent** charging (payment, proving a redelivered
`order.placed` never double-charges).

---

## Scaling a service

Because the user service is stateless and reached only through the gateway, you
can run several replicas — NGINX load-balances across them via Docker DNS:

```bash
docker compose up -d --scale user-service=3 user-service
docker compose ps          # see three user-service containers
```

No port conflicts, no config changes. This is the core "scalable" promise of the
architecture, demonstrable in one command.

---

## How the services were built (the pattern)

All six core services exist now. Every one was built from the **same template**,
documented step-by-step in
**[docs/adding-a-service.md](docs/adding-a-service.md)**:

1. `cp -r services/user services/<name>` and rename the package paths.
2. Replace `domain`, `migrations`, `repository`, `service`, `transport` with the
   new concepts.
3. Add storage (a `*-postgres` database, or reuse Redis) and a `*-service` block
   to `docker-compose.yml`.
4. Add the service's route block in
   [`gateway/nginx/conf.d/default.conf`](gateway/nginx/conf.d/default.conf).
5. Point the relevant storefront page at the new endpoint.

The services are also a catalogue of patterns to learn from:

- **`product`** — public reads + RBAC-guarded admin writes; an **event consumer**
  (`order.placed` → stock).
- **`cart`** — a non-Postgres (**Redis**) service; **synchronous REST** to another
  service (`pkg/restclient`).
- **`order`** — orchestration + **auth propagation** (forwarding the user's JWT)
  + saga coordination.
- **`payment`** — an **idempotent** event consumer that also publishes events.
- **`notification`** — a **pure subscriber**: no DB, no public API, decoupled from
  everyone.

Ideas to take it further: split into per-service Go modules/repos, move
`order.placed` consumers onto **JetStream durable consumers** for at-least-once
delivery, add a real Stripe integration behind the payment service's interface,
or build an admin dashboard against the existing APIs.

---

## Additional platform components (next steps)

The platform is deliberately a **lean core** so it runs comfortably on a laptop.
These production components are designed-for and easy to layer on:

- **Service discovery (Consul).** Today services find each other through Docker's
  built-in DNS (`user-service:8080`), which is enough for Compose. For dynamic,
  multi-host discovery and health-aware routing, add a Consul container and
  register services with it. The gateway already resolves upstreams at request
  time, which is the same pattern Consul-aware proxies use.
- **Centralized logging (ELK / Loki).** Every service already logs **structured
  JSON to stdout** with a request id. Point Filebeat/Promtail at the container
  logs to ship them into Elasticsearch/Loki and search across all services.
- **Metrics (Prometheus + Grafana).** `/healthz` and `/readyz` probes exist;
  add the `prometheus/client_golang` middleware to expose `/metrics`, scrape it
  with Prometheus, and chart it in Grafana.
- **CI/CD (GitHub Actions).** Each service builds independently
  (`go test ./...`, `docker build`), which maps cleanly onto a per-service
  pipeline.

Each of these is discussed with concrete steps in
[docs/architecture.md](docs/architecture.md#additional-components).

---

## What you'll learn

- Designing a system as **independent, single-responsibility services**.
- **Containerizing** Go and Node apps with small, secure multi-stage images.
- Orchestrating a multi-container system with **Docker Compose** (networks,
  volumes, healthchecks, `depends_on` conditions, scaling).
- Putting an **API gateway** in front and routing by path.
- **Synchronous** (REST) vs **asynchronous** (NATS events) communication and
  when to use each.
- Real-world **auth**: bcrypt, stateless JWTs, access/refresh rotation.
- **Clean architecture** that keeps business logic testable and infra swappable.
- The operational basics — **health probes, structured logging, graceful
  shutdown** — that make a service production-ready.

---

## Troubleshooting

| Symptom | Fix |
|--------|-----|
| `docker compose up` fails complaining about `JWT_SECRET` | You didn't create `.env`. Run `cp .env.example .env` and set a value. |
| Storefront loads but API calls fail | Make sure you opened **`http://localhost:8080`** (the gateway), not the Next dev server on `:3000`. |
| `user-service` restarts / "database not reachable" | Postgres is still booting; the service retries automatically. Check `docker compose logs user-postgres`. |
| Port `8080` already in use | Change the gateway's published port in `docker-compose.yml` (`"8081:80"`). |
| Frontend build fails on `next/font` | The Docker build needs network access to fetch the fonts once. Re-run with a connection. |

---

Built as a hands-on way to understand distributed systems by building one,
piece by piece. Start with [docs/getting-started.md](docs/getting-started.md).
