---
title: "Response Time Monitoring: Why 'Up but Slow' Is Still Down"
description: "Response time monitoring catches the slow degradations a simple up/down check misses. Learn to track latency, percentiles, and TTFB before users feel the pain."
date: 2026-06-26
author: pingdan
tags: [Performance, Monitoring]
keywords: [response time monitoring, latency monitoring, api latency, slow website, p95 p99 latency, ttfb, performance monitoring]
---

Your status page is green. Every check returned `200 OK`. And your support inbox is filling up with complaints that the app "feels broken." This is the blind spot of pure availability checks: a service that takes four seconds to respond is technically *up*, but to the user staring at a spinner, it is indistinguishable from down. Response time monitoring is how you close that gap.

## Availability Alone Tells You Half the Story

A binary up/down check answers one question: did the server respond? It says nothing about *how long* that took. Real degradations rarely announce themselves with a clean outage. They creep in as rising latency — a slow query here, a saturated connection pool there — until requests start timing out and your uptime check finally flips red. By then your users have been suffering for hours.

Availability and latency are two different signals, and both belong in your monitoring strategy. If you're still building out the fundamentals, start with [what is uptime monitoring](/blog/what-is-uptime-monitoring), then layer latency on top.

## What Response Time Actually Measures

"Response time" is not a single number — it's a sum of distinct phases, each with its own failure modes. When a monitor records total response time, it's stacking these segments:

- **DNS lookup** — resolving the hostname to an IP. Slow or flaky DNS adds latency before a single byte moves.
- **TCP connect** — establishing the socket. High values point to network distance or an overloaded host.
- **TLS handshake** — negotiating the encrypted session. Misconfigured ciphers or large cert chains show up here.
- **TTFB (Time To First Byte)** — the gap between request sent and first byte received. This is where *your application* lives: routing, database queries, business logic.
- **Content transfer** — streaming the full response body. Large payloads or slow downstream services inflate this.

When total latency climbs, breaking it into phases tells you *where*. A spike isolated to TTFB is almost always your code or database. A spike in TLS or connect time is infrastructure or network.

## Why Averages Lie

The single most common mistake in performance monitoring is watching the average. Averages smear outliers into invisibility. If 99 requests return in 50 ms and one takes 5 seconds, your average is ~100 ms — comfortable, and completely wrong about the user who waited five seconds.

Percentiles fix this. A percentile answers: "what's the worst experience for the fastest N% of requests?"

- **p50 (median)** — the typical request. Half are faster, half slower.
- **p95** — 95% of requests are at least this fast. The 1-in-20 slow path.
- **p99** — the tail. The 1-in-100 worst case, where retries, cold caches, and lock contention hide.

| Metric | What it shows | What it hides |
|--------|---------------|---------------|
| Average | A smeared blend | Every outlier that matters |
| p50 | Your typical user | The unlucky tail entirely |
| p95 | Where pain begins | The truly bad cases |
| p99 | Real worst-case latency | Nothing — this is the signal |

Your p99 is the experience of your most active users, because heavy users make more requests and are more likely to hit the slow tail. Optimize for p95 and p99, not the average.

## Mapping Latency to User Perception

Numbers need context. Here's a practical map from latency band to perception and action:

| Latency (TTFB) | User perception | Action |
|----------------|-----------------|--------|
| < 100 ms | Instant | Maintain |
| 100–300 ms | Snappy | Healthy baseline |
| 300 ms–1 s | Noticeable lag | Investigate trends |
| 1–3 s | Frustrating | Alert and triage |
| > 3 s | Likely abandoned | Page on-call |

These bands are starting points — tune them to your service and your users' tolerance.

## Setting Thresholds That Catch Degradation Early

The goal is to alert on *degradation*, not just failure. If you only alert when requests time out, you're alerting after the damage is done.

- **Baseline first.** Watch p95 and p99 over a normal week to learn what "good" looks like.
- **Alert on the percentile, not the average.** Trigger when p95 crosses a threshold above baseline — say, p95 > 800 ms sustained for 5 minutes.
- **Use sustained windows.** A single slow sample is noise; latency elevated across several consecutive checks is a trend.
- **Separate warning from critical.** Warn at degraded performance, page at user-impacting levels.

For broader alerting discipline, see [API monitoring best practices](/blog/api-monitoring-best-practices). And when you define what "acceptable" latency *is*, you're really defining an SLO — covered in [SLA, SLO, SLI explained](/blog/uptime-sla-slo-explained).

## Spotting Trends With Charts

A single latency reading is a data point; a chart is a story. Response-time charts reveal patterns a threshold alert can't:

- **Slow creep** — latency drifting up over days, the classic signature of a memory leak, growing table, or unbounded cache.
- **Time-of-day spikes** — peaks aligned with traffic, pointing to capacity limits.
- **Deploy correlation** — a step change right after a release. Overlay deploys on your latency chart and regressions become obvious.
- **Sawtooth patterns** — sharp drops after restarts followed by climbs, a tell for resource exhaustion.

Pair charts with deep assertions so you know the slow response is also *correct* — latency that returns wrong data is its own failure.

## Common Causes of Latency Creep

When the trend line points the wrong way, the usual suspects are:

- **Database queries** — missing indexes, N+1 patterns, table growth, or lock contention.
- **Connection pool exhaustion** — requests queuing for a free connection.
- **Downstream dependencies** — a third-party API or internal service degrading and dragging you with it.
- **Cold caches** — cache evictions or restarts forcing expensive recomputation.
- **Resource saturation** — CPU, memory, or I/O pressure as traffic grows past provisioned capacity.
- **Inefficient serialization** — payloads ballooning as your data model grows.

> Up is not the same as fast. A service that responds slowly enough is already failing its users — your monitoring should treat it that way.

Pingdan tracks response time on every check, breaks it down by phase, charts it over your full uptime history, and fires alerts through the channels your team watches the moment latency degrades — long before a timeout would.

[**Start monitoring free →**](/register)
