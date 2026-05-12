# pingdan

HTTP endpoint monitoring. Go API + in-process pinger, Next.js dashboard, Postgres.

## Architecture

- `api/` — Go service. Serves the REST API *and* runs the pinger as goroutines in the same process. Issues JWTs after Google/GitHub OAuth.
- `web/` — Next.js (App Router, TypeScript) dashboard. Talks to the Go API with a `Bearer` JWT stored in `localStorage`.
- `docker-compose.yml` — Postgres for local dev.

## Prerequisites

- Go 1.22+
- Node.js 20+
- Docker (for Postgres) or a local Postgres

## Setup

```sh
# 1. Start Postgres
docker compose up -d

# 2. API
cd api
cp .env.example .env
# fill in JWT_SECRET and at least one OAuth provider's client id/secret
go mod tidy
go run ./cmd/server

# 3. Dashboard (in another terminal)
cd web
cp .env.example .env.local
npm install
npm run dev
```

API at http://localhost:8080, dashboard at http://localhost:3000.

## OAuth redirect URIs

When configuring your OAuth apps, set the callback URLs to:

- Google: `http://localhost:8080/auth/google/callback`
- GitHub: `http://localhost:8080/auth/github/callback`

## API surface

| Method | Path                                            | Auth | Notes                                |
|--------|-------------------------------------------------|------|--------------------------------------|
| GET    | `/healthz`                                      | —    | Liveness                             |
| GET    | `/auth/{provider}/start`                        | —    | Begins OAuth (`google` or `github`)  |
| GET    | `/auth/{provider}/callback`                     | —    | Issues JWT, redirects to frontend    |
| GET    | `/me`                                           | JWT  | Current user                         |
| GET    | `/endpoints`                                    | JWT  | List your endpoints                  |
| POST   | `/endpoints`                                    | JWT  | Create                               |
| PUT    | `/endpoints/{id}`                               | JWT  | Update                               |
| DELETE | `/endpoints/{id}`                               | JWT  | Delete                               |
| GET    | `/alert-channels`                               | JWT  | List channels                        |
| POST   | `/alert-channels`                               | JWT  | Create (kind: `email` or `telegram`) |
| DELETE | `/alert-channels/{id}`                          | JWT  | Delete                               |
| POST   | `/endpoints/{id}/channels/{channelId}`          | JWT  | Attach channel                       |
| DELETE | `/endpoints/{id}/channels/{channelId}`          | JWT  | Detach channel                       |

## How pinging works

- Each enabled endpoint gets a goroutine with a `time.Ticker` at its `intervalSec`.
- Each tick fires an HTTP request with `timeoutSec` deadline; result is stored in `checks`.
- Endpoint flips to `down` after `failureThreshold` consecutive failures, back to `up` on the next success.
- On `up→down` or `down→up` transitions, all attached alert channels receive a notification.

## Alert channels

- **Email** — requires `SMTP_*` env vars in the API. Channel config: `{ "to": "you@example.com" }`.
- **Telegram** — requires `TELEGRAM_BOT_TOKEN`. Channel config: `{ "chatId": "123456789" }`. Get your chat ID by messaging the bot, then `https://api.telegram.org/bot<TOKEN>/getUpdates`.

Adding new channel types later means: a new `kind` value, a dispatcher branch in `internal/alerts/dispatcher.go`, and a UI option in `web/app/channels/page.tsx`.

## Production deploy (Caddy + Compose)

[docker-compose.prod.yml](docker-compose.prod.yml) runs Caddy, the API, the web app, and Postgres on a single host. Caddy handles Let's Encrypt automatically for two subdomains.

1. Point DNS for `app.example.com` and `api.example.com` at the server. Open ports 80 and 443.
2. Build and push images to a registry:
   ```sh
   # API
   docker build -t ghcr.io/you/pingdan-api:v0.1.0 ./api
   docker push ghcr.io/you/pingdan-api:v0.1.0
   # Web — NEXT_PUBLIC_API_URL is baked at build time
   docker build --build-arg NEXT_PUBLIC_API_URL=https://api.example.com \
       -t ghcr.io/you/pingdan-web:v0.1.0 ./web
   docker push ghcr.io/you/pingdan-web:v0.1.0
   ```
3. On the server:
   ```sh
   cp .env.prod.example .env
   # fill in domains, secrets, OAuth, image tags
   docker compose -f docker-compose.prod.yml --env-file .env up -d
   ```
4. Update OAuth callback URLs in Google/GitHub to `https://api.example.com/auth/{provider}/callback`.

Updating to a new release: bump the `API_IMAGE` / `WEB_IMAGE` tag in `.env` and `docker compose -f docker-compose.prod.yml --env-file .env up -d`.

## CI/CD (GitHub Actions)

Two workflows:

- [.github/workflows/ci.yml](.github/workflows/ci.yml) — runs on PRs against `main`. Builds the Go API and the Next.js app to verify nothing's broken.
- [.github/workflows/deploy.yml](.github/workflows/deploy.yml) — runs on push to `main`. Builds + pushes images to GHCR, then SSHes into the VPS, updates `.env`, and runs `docker compose pull && up -d`.

### One-time setup

1. **GHCR is automatic** — `GITHUB_TOKEN` already has `packages: write` via the workflow. No registry secrets needed.

2. **On the VPS** (one-time):
   ```sh
   mkdir -p /opt/pingdan && cd /opt/pingdan
   # create .env from .env.prod.example and fill in everything except API_IMAGE/WEB_IMAGE
   # (the deploy workflow will overwrite those two lines on each release)
   cp /path/to/.env.prod.example .env
   # ensure these two lines exist so `sed` can replace them:
   echo "API_IMAGE=" >> .env
   echo "WEB_IMAGE=" >> .env
   ```
   The workflow also `scp`s `docker-compose.prod.yml` and `Caddyfile` into this directory on every deploy.

3. **GitHub repository settings** → Settings → Environments → create `production`. Then under that environment add:

   **Secrets:**
   - `SSH_HOST` — VPS IP or hostname
   - `SSH_USERNAME` — VPS user (with docker access)
   - `SSH_PASSWORD` — VPS password
   - `SSH_PORT` — optional, defaults to 22
   - `DEPLOY_DIR` — e.g. `/opt/pingdan`

   **Variables:**
   - `API_DOMAIN` — e.g. `api.example.com` (used as a build-arg for the web image)

4. **Make the GHCR packages pullable from the VPS.** First push will create them as private. On the VPS, do one `docker login ghcr.io` to verify, or make the packages public in the GHCR UI if you don't mind.

### Rollback

Each build is tagged `sha-<7chars>`. To roll back, SSH in and:
```sh
cd /opt/pingdan
sed -i "s|^API_IMAGE=.*|API_IMAGE=ghcr.io/<owner>/pingdan-api:sha-abc1234|" .env
sed -i "s|^WEB_IMAGE=.*|WEB_IMAGE=ghcr.io/<owner>/pingdan-web:sha-abc1234|" .env
docker compose -f docker-compose.prod.yml --env-file .env up -d
```

## Production notes

- The pinger is in-process; the goroutine model handles low thousands of endpoints comfortably on one node. When you outgrow this, split the pinger into its own binary (the `endpoints.Store` and `alerts.Dispatcher` already work standalone).
- Migrations run on API startup via `goose`.
- Set `JWT_SECRET` to a long random value in production.
