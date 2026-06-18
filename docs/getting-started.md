# Getting started

This walks you from a clean machine to a running AURA platform, then to making
your first change. Commands are given for both **PowerShell** (Windows) and
**bash** (macOS/Linux/Git Bash).

---

## 1. Install the tools

You only strictly need **Docker** to run the whole stack. Go and Node are for
running pieces outside Docker (tests, frontend hot-reload).

### Docker Desktop (required)
- Download from <https://www.docker.com/products/docker-desktop/> and install.
- Launch it and wait until the whale icon says "Docker Desktop is running".
- Verify:
  ```bash
  docker --version
  docker compose version
  ```

### Go 1.23+ (optional — for tests / local builds)
- Windows: `winget install GoLang.Go`  (or download from <https://go.dev/dl/>)
- Verify: `go version`

### Node 20+ (optional — for frontend hot-reload)
- Windows: `winget install OpenJS.NodeJS.LTS`
- Verify: `node --version`

> After `winget install`, open a **new** terminal so the updated `PATH` is
> picked up.

---

## 2. Configure the environment

The stack reads secrets and settings from a `.env` file (Docker Compose loads it
automatically). Create it from the template:

```bash
# bash
cp .env.example .env
```
```powershell
# PowerShell
Copy-Item .env.example .env
```

Open `.env` and set a strong `JWT_SECRET`. Generate one:

```bash
# bash
openssl rand -base64 48
```
```powershell
# PowerShell
[Convert]::ToBase64String((1..48 | % { Get-Random -Maximum 256 }))
```

Paste it as the value of `JWT_SECRET`. The stack **refuses to start** if it's
left empty — that's deliberate, so misconfiguration fails loudly.

---

## 3. Start the platform

```bash
docker compose up --build
```

The first run takes a couple of minutes (it compiles Go and builds the Next.js
app). You'll know it's ready when you see the user service log
`http server listening` and `migrations applied`.

Open **<http://localhost:8080>** — you should see the AURA storefront.

To run it in the background: `docker compose up --build -d`, then
`docker compose logs -f` to tail logs.

---

## 4. Verify it works

**Through the UI:** click *Get started*, create an account, and you'll land on
your account page. Edit your display name and save — that's a real
`PATCH /users/me` round-trip.

**Through the API:**

```bash
curl -s -X POST http://localhost:8080/api/v1/auth/register \
  -H 'Content-Type: application/json' \
  -d '{"email":"you@example.com","name":"Your Name","password":"supersecret"}'
```

**Check service health:**

```bash
curl -s http://localhost:8080/healthz          # gateway
# readiness of the user service (DB + NATS reachable):
docker compose exec user-service /user healthcheck && echo OK
```

**Watch events flow:** open the NATS monitoring dashboard at
<http://localhost:8222> while you register users.

**Or run the whole flow at once** (register → login → profile → update →
refresh) with the included smoke test:

```bash
./scripts/smoke-test.sh                  # bash / Git Bash
```
```powershell
./scripts/smoke-test.ps1                 # PowerShell
```

---

## 5. Run the tests

```bash
go test ./...      # from the repo root
```

These exercise the user service's business logic with an in-memory fake
repository — no Docker or database required.

---

## 6. Develop the frontend with hot reload (optional)

Running the storefront outside Docker gives you instant hot reload. Point it at
the gateway for API calls:

```bash
cd web
npm install
# bash
NEXT_PUBLIC_API_BASE=http://localhost:8080 npm run dev
```
```powershell
# PowerShell
cd web
npm install
$env:NEXT_PUBLIC_API_BASE = "http://localhost:8080"; npm run dev
```

Now edit a component under `web/src` and the browser updates instantly, while
auth calls still hit the real user service through the gateway. (Keep
`docker compose up` running for the backend.)

---

## Common commands

```bash
docker compose up --build -d          # start in background
docker compose ps                     # list containers + health
docker compose logs -f user-service   # tail one service
docker compose down                   # stop (keeps data)
docker compose down -v                # stop and wipe the database volume
docker compose up -d --scale user-service=3 user-service   # 3 replicas
```

Or with the Makefile: `make up`, `make logs`, `make down`, `make test`,
`make scale-user`.

Next: read [architecture.md](architecture.md) to understand *why* it's shaped
this way, or jump to [adding-a-service.md](adding-a-service.md) to build the
Product service.
