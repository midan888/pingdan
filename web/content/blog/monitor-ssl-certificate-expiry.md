---
title: "Don't Let an Expired SSL Certificate Take You Down"
description: "SSL certificate monitoring catches expiry before it breaks production. Learn why certs still expire despite auto-renewal and how to alert weeks ahead."
date: 2026-06-27
author: pingdan
tags: [SSL, Monitoring]
keywords: [ssl certificate monitoring, certificate expiry, tls certificate expired, https monitoring, ssl expiry alert, certificate renewal]
---

An expired SSL certificate is one of the most embarrassing outages a team can ship, precisely because it is 100% preventable. Nothing crashed. No traffic spike took you down. No bad deploy slipped through review. The cert simply hit its `notAfter` date, and now every browser and API client on the planet is refusing to talk to your service. The fix takes minutes; the damage to trust and revenue does not.

The frustrating part is that this keeps happening to teams who *thought* they had it handled. Let's look at why TLS certificates still expire in production, what it costs you, and how a few minutes of monitoring turns a midnight fire drill into a calm calendar reminder.

## How TLS certificates expire

Every TLS certificate carries a validity window with a hard expiry timestamp. Once the current time passes `notAfter`, the certificate is invalid, full stop. Modern public certificates are short-lived by design — Let's Encrypt issues 90-day certs, and the CA/Browser Forum keeps pushing maximum lifetimes down. That's good for security, but it means renewal has to happen reliably, on a schedule, forever.

"Just use auto-renewal" is the standard answer. It is also where most expirations actually come from.

### Why it still happens despite auto-renewal

Auto-renewal is a background job, and background jobs fail silently:

- **The renewal cron broke and nobody noticed.** `certbot renew` runs from cron or a systemd timer. If the box was rebuilt, the timer was disabled, or a hook now exits non-zero, renewal stops — but the old cert keeps working *until the day it doesn't*. No alert fires because nothing is watching.
- **Wildcard and multi-domain (SAN) gaps.** A `*.example.com` wildcard does not cover `example.com` itself, and it does not cover `*.api.example.com`. New subdomains quietly fall outside the cert's coverage.
- **Internal services and intranets.** Certs on internal load balancers, admin panels, and service-to-service mTLS rarely have the same automation as your public edge. They are the most commonly forgotten.
- **Forgotten subdomains.** That marketing microsite, the old staging host, the webhook receiver someone spun up in 2023 — each has its own cert and its own clock.
- **Manual certs from a CA.** Anything issued by hand (EV certs, paid wildcards, certs pinned to a vendor) has no renewal automation at all. It lives entirely in someone's memory or a spreadsheet.

The common thread: the cert that expires is almost never the one you're watching closely. It's the edge case, the side project, the thing automation never covered.

## The user impact

When a cert expires, the failure is loud and absolute:

- **Scary full-page browser warnings.** `NET::ERR_CERT_DATE_INVALID` greets every visitor with an interstitial that screams "attackers may be trying to steal your information." Most users bounce immediately.
- **Blocked API clients.** SDKs, mobile apps, and server-to-server integrations validate the chain strictly. They don't show a warning — they just throw a TLS handshake error and your integrations go dark.
- **Dropped traffic and lost revenue.** Checkout flows, login, webhooks from payment providers — all of it stops. Unlike a soft degradation, there is no graceful fallback.
- **Reputation damage.** A cert error reads as "this company is sloppy," even to non-technical users, because the browser frames it as a security threat.

This is why an expired cert hurts more than an equivalent minute of downtime: it actively erodes trust, not just availability.

## How monitoring catches it early

The whole problem is *silence*. The cert works right up until it doesn't, so you need something actively watching and telling you with plenty of lead time. Good [uptime monitoring](/blog/what-is-uptime-monitoring) closes that gap on the certificate the same way it does for your endpoints:

- **Track days-remaining, not just up/down.** A check should read the certificate's expiry date and compute days left, then alert on a threshold — typically 30, 14, and 7 days out — so renewal happens on a weekday during business hours, never in a 2 a.m. scramble.
- **Validate HTTPS endpoint health end to end.** Confirm the handshake completes, the chain is trusted, and the hostname matches. A cert can be present and valid but served with a broken chain — see below.
- **Assert on response content, not just connectivity.** Pingdan's deep assertions verify [HTTP status codes](/blog/http-status-codes-for-monitoring), response headers, and body content over HTTPS, so you catch a degraded TLS endpoint that still technically answers.
- **Route alerts where people see them.** Email and Telegram alerts mean the warning reaches a human days before the deadline instead of arriving as a customer support ticket after the fact.

| Days remaining | Recommended action |
| --- | --- |
| 30 days | First alert — confirm auto-renewal is scheduled and healthy |
| 14 days | Investigate if not yet renewed; check the renewal logs |
| 7 days | Treat as urgent; renew manually if automation is stuck |
| 0 days | Outage in progress — emergency renewal and incident review |

## A practical pre-renewal checklist

Before and during renewal, walk through this:

- [ ] Confirm the renewal job (certbot, ACME client, or vendor automation) ran successfully this cycle.
- [ ] Verify the new cert's expiry date actually advanced — don't trust "exit code 0" alone.
- [ ] Reload or restart the web server / load balancer so the new cert is served (a renewed file on disk that isn't loaded changes nothing).
- [ ] Check the served chain from outside your network, not just locally.
- [ ] Confirm every SAN and subdomain you depend on is covered by the new cert.
- [ ] Re-verify internal and non-production hosts on the same cadence.
- [ ] Make sure your monitoring is pointed at the actual public hostname.

## Related failure modes to watch

Expiry is the headline, but adjacent TLS issues fail in similar ways:

- **Mixed content.** An HTTPS page loading scripts or images over HTTP gets blocked by browsers, breaking functionality even with a valid cert.
- **HSTS.** With `Strict-Transport-Security` set, browsers refuse to let users click through a cert error at all. Great for security, brutal if your cert lapses — there is no bypass.
- **Incomplete chain.** If your server omits an intermediate certificate, some clients (especially older ones and non-browser HTTP libraries) reject the connection even though browsers may paper over it.

Folding these checks into your [API monitoring best practices](/blog/api-monitoring-best-practices) means TLS health is verified continuously, not just on renewal day.

> An expired certificate is the rare outage you can schedule yourself out of entirely. The only thing standing between you and a preventable incident is something that watches the clock so you don't have to.

[**Start monitoring free →**](/register)
