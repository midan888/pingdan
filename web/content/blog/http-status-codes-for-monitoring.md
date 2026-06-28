---
title: "HTTP Status Codes for Monitoring: Which Ones Mean You're Down"
description: "A practical guide to HTTP status codes monitoring: how to tell 4xx vs 5xx apart, which codes mean you're really down, and how a monitor should react."
date: 2026-06-14
author: pingdan
tags: [HTTP, Monitoring]
keywords: [http status codes monitoring, 500 error, 503 service unavailable, 4xx vs 5xx, what status code means down, monitor http errors]
---

Your monitor just paged you for a `403`. Is the service down, or did someone rotate an API key? Status codes are the first signal a monitor sees, but treating them all as "up" or "down" is how you end up with alert fatigue and missed outages.

This guide breaks down what each class of HTTP status code actually means for monitoring, and how a check should react to each one.

## The quick reference

Here is the short version. The right reaction depends as much on the endpoint as the code itself, but this is a sane default for most public APIs and web apps.

| Code | Meaning | Typical monitoring action |
|------|---------|---------------------------|
| 200 / 204 | OK | Pass — but verify the body too |
| 301 / 302 | Redirect | Follow once; alert on loops or chains |
| 401 / 403 | Unauthorized / Forbidden | Usually *your* check is misconfigured |
| 404 | Not found | Down if the endpoint should exist |
| 429 | Too many requests | Back off; not an outage |
| 500 | Internal server error | Down — real failure |
| 502 / 504 | Bad gateway / Gateway timeout | Down — upstream or proxy failing |
| 503 | Service unavailable | Down or maintenance — check intent |

## 2xx: success that can still be broken

A `200 OK` is the happy path, but it only tells you the server *responded*, not that it responded *correctly*. Plenty of broken apps return `200` with a stack trace, an empty payload, an error page, or a JSON body that says `{"error": "database unavailable"}`.

This is why status-only monitoring is a trap. A real check needs to assert on the response itself:

```http
GET /api/v1/orders HTTP/1.1
Host: api.example.com

HTTP/1.1 200 OK
Content-Type: application/json

{"status": "degraded", "orders": null}
```

That `200` should fail your check. The fix is to layer assertions on top of the status code — match a header, a substring in the body, or a value at a [JSON path](/blog/json-path-assertions-guide) like `$.status == "ok"`. Status code plus body assertions is the difference between "the server is on" and "the feature works."

## 3xx: redirects, chains, and loops

Redirects are normal — `301` for permanent moves, `302`/`307` for temporary ones. A monitor should generally follow a single redirect and validate the final destination.

Two things to watch:

- **Redirect chains.** `http → https → www → /home` adds latency and can hide a misconfiguration. If you expected a `200` and got three hops to get there, that is worth surfacing.
- **Redirect loops.** A misconfigured load balancer or auth gateway can bounce a request between two URLs forever. Cap the number of redirects your check follows (5 is reasonable) and treat exhaustion as a failure.

If your monitor follows redirects silently, you can miss the day someone accidentally points production at a `302` to a login page.

## 4xx: client errors — usually your fault, not the server's

The `4xx` class means the *request* was the problem. For monitoring, that almost always means your check is misconfigured rather than the service being down.

- **401 Unauthorized / 403 Forbidden** — Your token expired, an API key rotated, or an IP allowlist changed. The service is probably fine; your monitor's credentials are stale. Alert, but label it as a config issue, not an outage.
- **404 Not Found** — Context-dependent. A `404` on a route that *should* exist is a real failure (someone shipped a bad deploy). A `404` you expected (probing a deliberately-missing path) is a pass.
- **429 Too Many Requests** — You are being rate-limited. This is not an outage. Your monitor should respect `Retry-After`, back off, and avoid hammering the endpoint into a worse state.

The key insight: a healthy `4xx` check is one where you tell the monitor which code you *expect*. Asserting "expect `401` on this protected route" is a valid, useful check — it confirms auth is still enforced.

## 5xx: the real down signals

The `5xx` class is where "down" usually lives. The server accepted the request and failed to fulfill it.

- **500 Internal Server Error** — An unhandled exception. Real, server-side breakage. Treat as down.
- **502 Bad Gateway** — A proxy or load balancer got an invalid response from upstream. Often means app instances crashed or are not registered. Down.
- **503 Service Unavailable** — The server is up but refusing work: overloaded, deploying, or in maintenance. This is genuinely down for users — though if it is *planned* maintenance, you may want a maintenance window rather than a page.
- **504 Gateway Timeout** — The upstream did not respond in time. Down, and usually a symptom of a slow dependency or a saturated backend.

For all of these, the right move is to confirm before paging. A single `5xx` can be a transient blip; several in a row across consecutive checks is an outage. Requiring 2-3 consecutive failures before alerting cuts noise dramatically — more on that in [API monitoring best practices](/blog/api-monitoring-best-practices).

## No status code at all: timeouts and connection errors

Sometimes there is no status code, and these are often the *worst* failures.

- **Connection refused** — Nothing is listening. The process is dead or the port is wrong.
- **DNS failure** — The hostname will not resolve. Could be an expired domain or a broken DNS change.
- **TLS errors** — An expired or mismatched certificate. Browsers will block users entirely.
- **Timeout** — No response within your window. The server may be alive but pathologically slow, which for users is indistinguishable from down.

Set an explicit timeout (commonly 10-30 seconds) and treat exceeding it as a failure. A slow `200` that arrives after 45 seconds is not a success. Tracking response time alongside status is core to [uptime monitoring](/blog/what-is-uptime-monitoring) — a latency spike often precedes a full `5xx` outage by minutes.

## Putting it together

A robust check is a decision tree, not a binary:

- Did we get *any* response? No → connection/timeout failure.
- Is the status the one we expect? `200`, `401`, whatever you declared — not just "2xx good, 4xx bad."
- Does the body back it up? Assert on headers, content, and JSON paths.
- Is this a one-off or a trend? Require consecutive failures before paging.

> The status code tells you the server answered. Your assertions tell you it answered *correctly*. Monitoring only the first is how outages slip through a sea of green checkmarks.

[**Start monitoring free →**](/register)
