# Adding a new service (worked example: Product Catalog)

The User service is built to be a **template**. This guide turns it into the
**Product Catalog** service step by step. The same recipe applies to Cart,
Order, Payment, and Notification.

Time budget: ~1–2 hours the first time, much less afterwards.

> ✅ **This service is already implemented** in `services/product` — it was built
> by following this exact guide. Read it alongside the code as a worked example,
> then apply the same steps to Cart/Order/Payment/Notification.
>
> Two things the generic guide simplifies, now that the platform has shared
> infrastructure:
> - The Postgres pool + migration runner is **shared** in [`pkg/db`](../pkg/db/db.go).
>   A new service's `repository/postgres` package only contains its own
>   `*_repo.go` queries; `main.go` calls `db.NewPool` / `db.Migrate`.
> - Role-based auth is shared: `pkg/auth` defines `RoleAdmin`, the user service
>   stamps the `role` claim into JWTs, and `middleware.RequireRole(auth.RoleAdmin)`
>   guards admin endpoints (see Step 4).

---

## The mental model

Every service is the same six layers you already have in `services/user`:

```
domain/        entities + errors            (Product, ErrProductNotFound)
repository/    Postgres adapter + migrations (products table)
service/       business logic               (list, get, create, adjust stock)
transport/http handlers + DTOs + router      (GET /products, …)
events/        published event payloads      (product.created)
cmd/server/    wiring + graceful shutdown
```

Plus three integration points outside the service:
`docker-compose.yml` (container + DB), the **gateway route**, and the
**storefront** call.

---

## Step 1 — Copy the template

```bash
cp -r services/user services/product
```

Then find-and-replace the import paths and names inside `services/product`:

- `services/user` → `services/product` (import paths)
- `user-service` → `product-service` (default `SERVICE_NAME`)
- package-level names `User`/`user` → `Product`/`product`

> Because everything lives in one Go module
> (`github.com/aura-commerce/platform`), the import path is just the folder path
> — no new `go.mod` needed.

---

## Step 2 — Model the domain

Edit `services/product/internal/domain/product.go`:

```go
type Product struct {
    ID          string
    Name        string
    Slug        string
    Description string
    Category    string
    PriceCents  int      // money is always integer cents
    Stock       int
    CreatedAt   time.Time
    UpdatedAt   time.Time
}

var (
    ErrProductNotFound = errors.New("product not found")
    ErrSlugTaken       = errors.New("slug already exists")
    ErrOutOfStock      = errors.New("insufficient stock")
)
```

---

## Step 3 — Migration

Replace `services/product/migrations/0001_init.sql`:

```sql
CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS products (
    id          UUID PRIMARY KEY,
    slug        TEXT        NOT NULL UNIQUE,
    name        TEXT        NOT NULL,
    description TEXT        NOT NULL DEFAULT '',
    category    TEXT        NOT NULL,
    price_cents INTEGER     NOT NULL CHECK (price_cents >= 0),
    stock       INTEGER     NOT NULL DEFAULT 0 CHECK (stock >= 0),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_products_category ON products (category);
```

The embedded migration runner (`pkg`/`postgres.Migrate`) applies it on startup —
nothing else to wire.

---

## Step 4 — Repository, service, transport

- **repository/postgres/product_repo.go** — same shape as `user_repo.go`:
  `Create`, `GetByID`, `GetBySlug`, `List(filter)`, `AdjustStock`. Map the
  unique-violation SQLSTATE `23505` → `domain.ErrSlugTaken`.
- **service/service.go** — business rules: generate the slug, validate price,
  decrement stock (returning `ErrOutOfStock`), publish `product.created`.
- **transport/http** — DTOs + handlers + router. Public reads, admin-only writes:

  ```go
  r.Route("/api/v1/products", func(p chi.Router) {
      p.Get("/", h.List)          // public
      p.Get("/{slug}", h.GetOne)  // public
      p.Group(func(admin chi.Router) {
          admin.Use(middleware.Authenticated(tokens))        // require a token…
          admin.Use(middleware.RequireRole(auth.RoleAdmin))  // …with role=admin
          admin.Post("/", h.Create)
          admin.Post("/{id}/stock", h.AdjustStock)
      })
  })
  ```

  `RequireRole` (in `pkg/middleware`) reads the `role` claim and 403s anyone who
  isn't an admin. It's shared, so every service guards admin routes the same way.

---

## Step 5 — Add the database and service to Compose

In `docker-compose.yml`:

```yaml
  product-service:
    build:
      context: .
      dockerfile: services/product/Dockerfile
    environment:
      SERVICE_NAME: product-service
      HTTP_ADDR: ":8080"
      DATABASE_URL: postgres://aura:${PRODUCT_DB_PASSWORD:-aura_dev}@product-postgres:5432/products?sslmode=disable
      NATS_URL: nats://nats:4222
      JWT_SECRET: ${JWT_SECRET:?set JWT_SECRET}
    depends_on:
      product-postgres: { condition: service_healthy }
      nats: { condition: service_started }
    networks: [aura]

  product-postgres:
    image: postgres:16-alpine
    environment:
      POSTGRES_USER: aura
      POSTGRES_PASSWORD: ${PRODUCT_DB_PASSWORD:-aura_dev}
      POSTGRES_DB: products
    volumes:
      - product_pgdata:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U aura -d products"]
      interval: 5s
      timeout: 3s
      retries: 10
    networks: [aura]
```

Add `product_pgdata:` under the top-level `volumes:`. Note it's a **separate
database** — that's the database-per-service rule in action.

---

## Step 6 — Open the gateway route

In `gateway/nginx/conf.d/default.conf`, uncomment / add:

```nginx
location /api/v1/products/ {
    set $product_service http://product-service:8080;
    proxy_pass $product_service;
}
```

Reload by recreating the gateway: `docker compose up -d gateway`.

---

## Step 7 — Point the storefront at it

In `web/src/lib`, add a `products-api.ts` that fetches `/api/v1/products` and
returns the same `Product` shape the UI already renders. Then change
`web/src/app/products/page.tsx` to load from the API instead of the static
`products` array. The components don't change — they already render from that
shape.

---

## Step 8 — Run it

```bash
docker compose up --build -d product-service product-postgres gateway
curl -s http://localhost:8080/api/v1/products | jq
```

You now have a second independent service, with its own database, reachable
through the gateway, consumed by the storefront — built entirely by following
the User service's pattern.

---

## The rest of the services (all built — study them)

Cart, Order, Payment and Notification were all built with this recipe and are in
the repo now. Each one adds a pattern worth studying:

| Service | Storage | Pattern it demonstrates |
|--------|---------|-------------------------|
| **`cart`** | **Redis** | A non-Postgres service; **sync REST** to the product service via [`pkg/restclient`](../pkg/restclient/restclient.go) to snapshot price/name |
| **`order`** | Postgres | Orchestration; **auth propagation** (forwards the user's JWT to the cart service); emits `order.placed`; subscribes to payment results |
| **`payment`** | Postgres | **Idempotent** event consumer (one payment per `order_id`); simulated charge behind a swappable `charge()` method; emits `payment.succeeded` |
| **`notification`** | none | A **pure subscriber**: `bus.Subscribe(...)` for three events, no DB, no business API — added with **zero changes** to any other service |

To extend the saga, new event consumers go in a service's `internal/subscriber`
package and are wired in its `main.go` with `subscriber.Register(bus, svc, log)`
— exactly as the product/order/payment/notification services already do.
