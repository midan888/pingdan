---
name: verify
description: Build, run, and drive pingdan (Go API + Next.js web) locally to verify a change end-to-end.
---

# Verifying pingdan changes

Two apps: Go API (`api/`, chi + pgx, port :8080 by default) and Next.js web (`web/`, App Router).

## Isolated environment (does not touch dev data)

```bash
# 1. Postgres via compose, plus a throwaway DB (migrations auto-run at API startup)
docker compose up -d postgres
docker compose exec -T postgres psql -U pingdan -c "CREATE DATABASE pingdan_verify;"

# 2. API — env vars beat api/.env (godotenv doesn't override existing env)
cd api && go build -o /tmp/pingdan-api ./cmd/server
HTTP_ADDR=:8099 FRONTEND_URL=http://localhost:3100 \
  DATABASE_URL="postgres://pingdan:pingdan@localhost:5432/pingdan_verify?sslmode=disable" \
  JWT_SECRET=verify-secret /tmp/pingdan-api

# 3. Web
cd web && NEXT_PUBLIC_API_URL=http://localhost:8099 npx next dev -p 3100
```

**Gotcha:** CORS. The API only allows the origin in `FRONTEND_URL` (default
`http://localhost:3000`). If the web dev server runs on another port, every
browser fetch fails silently — pages show empty data. Always set
`FRONTEND_URL` to match the web port.

## Getting a token / seeding data

```bash
curl -s -X POST :8099/auth/email/register -H 'Content-Type: application/json' \
  -d '{"email":"a@test.local","password":"secret1234","name":"A"}'   # → {"token": ...}
curl -s -X POST :8099/endpoints -H "Authorization: Bearer $TOK" \
  -H 'Content-Type: application/json' -d '{"name":"e1","url":"https://example.com"}'
```

The web app stores the JWT in `localStorage` under key `pingdan_token`. To
drive pages with Playwright: goto `/login`, `localStorage.setItem("pingdan_token", tok)`,
then goto the target page. `playwright-core` + `npx playwright-core install chromium`
works from any scratch dir.

Admin routes (`/admin/*`) are gated by the `ADMIN_EMAILS` env var
(comma-separated, case-insensitive); `/me` returns `isAdmin`.

## Cleanup

Stop both servers, then:
```bash
docker compose exec -T postgres psql -U pingdan -c "DROP DATABASE pingdan_verify;"
docker compose stop postgres   # if it wasn't running before
```
