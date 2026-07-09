# Alert channel integrations — implementation plan

Goal: grow alert channels from the current two (email via Resend, Telegram) to a full set:
Slack, Discord, Microsoft Teams, generic webhook, PagerDuty, ntfy, Pushover, Twilio SMS, Opsgenie.

Each phase below is self-contained and sized to be implemented in its own session.
Do them in order — Phase 0 is a prerequisite for everything else.

**Status**

- [ ] Phase 0 — Dispatcher + channels-form refactor (foundation)
- [ ] Phase 1 — Slack, Discord, Teams, generic webhook
- [ ] Phase 2 — PagerDuty, ntfy, Pushover
- [ ] Phase 3 — Twilio SMS, Opsgenie
- [ ] Cross-cutting follow-ups (docs, marketing pages)

---

## Architecture context (read first in every session)

- Backend is Go (`api/`), frontend is Next.js (`web/`).
- Alert dispatch lives in `api/internal/alerts/dispatcher.go`. `Dispatcher` has `Notify`
  (down/recovered), `NotifySSL` (cert expiry), and `SendTest`. Each currently repeats a
  `switch kind` over `email` / `telegram`.
- HTTP handlers: `api/internal/http/alert_handlers.go`. Kind allowlist is hardcoded twice
  (`create` and `test` handlers).
- Channel rows: `alert_channels` table — `kind TEXT`, `config JSONB`. **No migration is
  needed to add kinds.** Endpoints attach via `endpoint_alert_channels`.
- Wiring: `api/cmd/server/main.go` builds the `Dispatcher` from `api/internal/config/config.go`
  (env vars).
- UI: `web/app/channels/page.tsx` — a `KINDS` array plus hardcoded email/telegram branches in
  `validate()`, `configFor()`, and the list-item `target` display. Types in `web/lib/api.ts`.

Conventions: every new sender is "POST JSON, treat status ≥ 300 as error, log a truncated
response body" — exactly like the existing `sendTelegram`. Per-user credentials go in the
channel `config` JSONB; app-level credentials (one per deployment) go in env vars via
`config.go`.

---

## Phase 0 — Refactor dispatcher and channels form (foundation)

No new integrations; makes Phases 1–3 mechanical. Backend and frontend parts are independent.

### Backend (`api/internal/alerts/`, `api/internal/http/`)

1. Add an `Alert` struct in the `alerts` package carrying structured data (needed later by
   PagerDuty dedup/resolve and the generic webhook payload):
   - `Event`: `endpoint.down` | `endpoint.recovered` | `ssl.expiring` | `test`
   - `Endpoint`: id, name, url
   - `Check`: status code, error, checked-at (nullable — absent for SSL/test)
   - `SSL`: days left, expires-at (nullable)
   - `Subject`, `Body`: pre-rendered text for text-only channels (current `renderMessage`
     / `renderSSLMessage` output)
2. Replace the three `switch kind` blocks with a single
   `send(ctx, kind string, cfg []byte, a Alert) error` dispatching through a
   `map[string]func(...)` of senders. `Notify`, `NotifySSL`, `SendTest` build an `Alert`
   and delegate.
3. Add a shared `postJSON(ctx, url string, headers map[string]string, payload any) error`
   helper with a short timeout; rewrite `sendTelegram` (and the Resend call in `sendEmail`)
   on top of it.
4. Export `alerts.ValidKinds` / `alerts.IsValidKind(kind string) bool`; use it in
   `alert_handlers.go` `create` and `test` instead of the hardcoded string comparisons.
5. Make senders unit-testable: API base URLs (Telegram, Resend, later PagerDuty etc.)
   overridable on `Dispatcher` (or inject `*http.Client`), so tests can point at an
   `httptest.Server`. Add tests asserting payload shape for email + telegram, matching the
   style of `api/internal/assertions/assertions_test.go`.

### Frontend (`web/app/channels/page.tsx`)

6. Generalize `KINDS` so each kind declares its own field list:
   `fields: [{ key, label, placeholder, hint, inputMode?, validate?, optional? }]`.
   Derive `validate()`, `configFor()`, the form inputs, and the list-item `target` display
   from this data instead of email/telegram conditionals. (Generic webhook in Phase 1 needs
   two fields — url + optional secret — which the current single-value form can't express.)
7. Extend the `Kind` union / `AlertChannel` type in `web/lib/api.ts` as needed; check
   `web/components/EndpointForm.tsx` and the endpoint detail page for hardcoded kind
   rendering.

**Acceptance:** behavior identical to before (email + telegram add/test/notify all work);
adding a new kind now touches only: one sender func + registry entry + `ValidKinds` +
one `KINDS` entry in the UI.

---

## Phase 1 — Slack, Discord, Teams, generic webhook

All four are per-user webhook URLs stored in channel `config` — zero new env vars.
For each: config struct + sender registered in the Phase 0 map + `ValidKinds` entry +
`KINDS` UI entry with validation.

1. **Slack** — config `{ "webhookUrl" }`. POST `{"text": subject + "\n" + body}` to the URL.
   UI validates prefix `https://hooks.slack.com/`.
2. **Discord** — config `{ "webhookUrl" }`. POST `{"content": ...}`, truncated to 2000 chars
   (Discord hard limit). UI validates prefix `https://discord.com/api/webhooks/`
   (also accept `discordapp.com`).
3. **Microsoft Teams** — config `{ "webhookUrl" }`. Target Power Automate workflow webhooks
   (classic O365 connectors are retired). Payload is an Adaptive Card envelope:
   `{"type":"message","attachments":[{"contentType":"application/vnd.microsoft.card.adaptive","content":{...}}]}`.
   Card: title = subject, body = alert lines as facts. URLs live on `*.logic.azure.com`;
   UI just requires https.
4. **Generic webhook** — config `{ "url", "secret"? }`. POST the full structured `Alert` as
   JSON. If `secret` is set, sign: `X-Pingdan-Signature: sha256=<hex hmac-sha256 of body>`.
   Two form fields (url required, secret optional) — uses the Phase 0 multi-field form.
5. **SSRF hardening** (decide once, applies to all user-supplied URLs): require http/https,
   short client timeout, no redirects following to non-https. Optional stretch: resolve host
   and reject private/loopback ranges.
6. Unit tests per sender via `httptest.Server` (payload shape, signature correctness,
   Discord truncation).

**Acceptance:** all four kinds can be added, tested ("Send test"), attached to an endpoint,
and fire on down/recovered/SSL. Generic-webhook signature verifiable with a known secret.

---

## Phase 2 — PagerDuty, ntfy, Pushover

1. **PagerDuty** — config `{ "routingKey" }`. Events API v2:
   `POST https://events.pagerduty.com/v2/enqueue` with
   `{routing_key, event_action, dedup_key, payload: {summary, source, severity}}`.
   Mapping:
   - `endpoint.down` → `trigger`, `dedup_key: "pingdan-endpoint-<id>"`, severity `critical`
   - `endpoint.recovered` → `resolve`, same dedup key (incident auto-closes)
   - `ssl.expiring` → `trigger`, `dedup_key: "pingdan-ssl-<id>"`, severity `warning`
   - `test` → `trigger` immediately followed by `resolve` (no lingering incident)
2. **ntfy** — config `{ "topic", "server"?, "accessToken"? }`, server default
   `https://ntfy.sh`. POST plain-text body to `<server>/<topic>` with headers `Title`
   (subject), `Priority` (`urgent` on down, `default` otherwise), and
   `Authorization: Bearer <token>` when set.
3. **Pushover** — needs an app-level token: register a "pingdan" app once, add
   `PUSHOVER_APP_TOKEN` to `config.go`, `main.go` wiring, and `docker-compose.prod.yml`.
   Channel config `{ "userKey" }`. POST form-encoded to
   `https://api.pushover.net/1/messages.json` (`token`, `user`, `title`, `message`).
   If env var unset, sender errors like the existing Telegram-token check; optionally hide
   the kind in the UI via a capabilities flag.
4. Unit tests: PagerDuty trigger/resolve mapping per event type, ntfy headers, Pushover form
   encoding.

**Acceptance:** PagerDuty incident opens on down and auto-resolves on recovery; ntfy and
Pushover deliver to a phone; "Send test" works for all three.

---

## Phase 3 — Twilio SMS, Opsgenie

1. **Twilio SMS** — env vars `TWILIO_ACCOUNT_SID`, `TWILIO_AUTH_TOKEN`, `TWILIO_FROM`
   (config.go + main.go + docker-compose.prod.yml). Channel config `{ "to" }` — E.164,
   validate in UI (`^\+[1-9]\d{6,14}$`). POST form-encoded with HTTP basic auth to
   `https://api.twilio.com/2010-04-01/Accounts/<sid>/Messages.json`. Keep the body terse
   (endpoint name + state only) — SMS costs per message. Plan-gating is a later product
   decision; plumbing doesn't depend on it.
2. **Opsgenie** — config `{ "apiKey", "region": "us" | "eu" }` (region selects
   `api.opsgenie.com` vs `api.eu.opsgenie.com`). `Authorization: GenieKey <key>`.
   Create alert on down with `alias: "pingdan-endpoint-<id>"` as dedup; close the alert
   via the alias on recovery. Mirror the PagerDuty event mapping, including test behavior.
3. **Deliberately skipped:** Better Stack, Grafana OnCall — both ingest generic webhooks,
   so Phase 1's webhook channel covers them. Add docs recipes instead (see cross-cutting).

**Acceptance:** SMS arrives on down/recovered; Opsgenie alert opens and closes with state.

---

## Cross-cutting follow-ups (last session, or fold into Phase 1)

- **Docs page** (`web/app/docs/page.tsx`): setup instructions per channel (how to create a
  Slack/Discord/Teams webhook, PagerDuty routing key, ntfy topic, etc.); document the
  generic-webhook JSON payload and signature verification; recipes for Better Stack /
  Grafana OnCall via generic webhook.
- **Marketing** (`web/app/features/page.tsx`, landing page, pricing): update integration
  lists that currently say email + Telegram.
- **Env template / README**: document new optional env vars (`PUSHOVER_APP_TOKEN`,
  `TWILIO_*`).
- Consider a `GET /capabilities` (or fold into an existing endpoint) exposing which
  env-var-backed kinds are configured, so the UI hides Pushover/Twilio/Telegram when the
  deployment lacks credentials.
