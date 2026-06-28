---
title: "How Often Should You Check Your Endpoints? Picking a Monitor Interval"
description: "Your monitor check interval is a trade-off between fast detection and noise. Here's how to pick the right frequency for marketing sites, APIs, and crons."
date: 2026-06-18
author: pingdan
tags: [Monitoring, Best Practices]
keywords: [monitor check interval, how often to monitor, uptime check frequency, monitoring interval, 1 minute checks, detection time]
---

You set up a monitor, and the first decision the tool asks for is deceptively simple: how often should it run? Every minute? Every five? Every hour? It feels like a slider you can drag to "more is better" and forget about. It isn't. Your **monitor check interval** is one of the most consequential settings you'll touch, because it silently caps how fast you can ever find out something is broken — and how much noise you sign up for in return.

## The core trade-off

A shorter interval buys you faster detection. It also costs you more: more requests against your endpoints, more load on rate-limited or metered APIs, and more opportunities for a single flaky response to wake you up at 3 a.m. A longer interval is cheap and quiet, but it means an outage can run for ages before anyone notices.

There's no universally correct number. The right interval depends on what the endpoint *does* and what a minute of downtime actually costs you. A static marketing page and a checkout API don't deserve the same treatment.

## Interval and mean-time-to-detect

The cleanest way to reason about this is **mean-time-to-detect (MTTD)** — how long, on average, a problem exists before your monitoring catches it.

If you check on a fixed interval, an outage can begin at any point between two checks. On average it starts halfway through the gap, so your expected detection time is roughly *half the interval*, plus the time your failure threshold adds on top (more on that below). A 5-minute interval means you'll typically learn about an outage about 2.5 minutes in, and at worst close to 5 minutes in.

| Check interval | Avg detection time | Good for |
| --- | --- | --- |
| 30 seconds | ~15 seconds | Revenue paths, payment & auth APIs, critical real-time services |
| 1 minute | ~30 seconds | Core product APIs, login, anything customer-facing |
| 5 minutes | ~2.5 minutes | Marketing sites, blogs, dashboards, secondary services |
| 15 minutes | ~7.5 minutes | Internal tools, staging, low-traffic admin panels |
| 1 hour | ~30 minutes | Cron jobs, scheduled exports, nightly batch endpoints |

These detection times are the *floor*. They assume an instant alert on the first failed check. Add a failure threshold and the real number climbs.

## Recommended intervals by tier

Rather than picking one global setting, group your endpoints into tiers.

### Marketing & content (5 minutes)

A landing page being down for a couple of minutes is annoying, not catastrophic. Checking every 5 minutes keeps cost and noise low while still catching real outages well before most visitors complain. If you're new to all of this, start here and read [what is uptime monitoring](/blog/what-is-uptime-monitoring) for the fundamentals.

### Core API & revenue paths (30–60 seconds)

This is where money lives: checkout, payments, sign-in, the API your customers integrate against. Every minute of downtime here is lost revenue and lost trust, so sub-minute checks earn their keep. The math is blunt — at a 5-minute interval you might not even alert before an outage has already cost you several minutes of failed transactions, but at 30 seconds you're reacting almost immediately. Pair tight intervals with [deep assertions](/blog/api-monitoring-best-practices) so you catch a `200 OK` that's actually returning a broken payload, not just a connection failure.

### Internal crons & batch jobs (15 minutes to 1 hour)

A nightly export or a job that runs every few hours doesn't need second-by-second watching. Match the interval to the cadence of the work. Checking a once-an-hour cron every 30 seconds just hammers it for no benefit — you'll only ever learn something new once an hour anyway.

## Why sub-minute matters for revenue paths

For most pages, the difference between 1-minute and 5-minute checks is academic. For the endpoints that take money, it's not. Outages on payment or auth flows convert directly into failed orders and support tickets, and the cost is roughly linear with duration. Shaving detection from ~2.5 minutes down to ~15 seconds can be the difference between a non-event and a string of angry refunds. Reserve your most aggressive intervals for the handful of routes where speed genuinely pays for itself.

## Interval and failure thresholds work together

Here's the subtlety most people miss: you almost never want to alert on a *single* failed check. The internet is noisy. One timeout doesn't mean an outage — it might be a transient network blip or a single bad node.

So you set a threshold: alert after, say, 2 or 3 consecutive failures. That's good hygiene, but it multiplies your detection time. With a 1-minute interval and a 3-strike threshold, your real worst-case detection is closer to three minutes, not one.

This is the lever to balance:

- **Short interval + low threshold** = fastest detection, most false positives.
- **Short interval + higher threshold** = the sweet spot for critical endpoints — you confirm quickly *and* filter noise.
- **Long interval + high threshold** = very quiet, very slow. Fine for low-stakes targets.

For a revenue endpoint, a 30-second interval with a 2-failure threshold confirms a real problem in about a minute while ignoring one-off blips. Tuning thresholds is also the single best defense against pager burnout — see [reducing alert fatigue](/blog/reduce-alert-fatigue) for how to keep alerts meaningful.

## Don't hammer rate-limited or expensive endpoints

A faster interval isn't free for the *target* either. Some endpoints are rate-limited, some are metered per request, and some kick off real work — a database query, a third-party API call, an LLM completion. Checking those every 30 seconds can blow through a quota, run up a bill, or trip the very rate limits you're trying to monitor.

A few ways to stay polite:

- Point health checks at a cheap, dedicated `/health` route rather than an expensive business endpoint.
- Back off the interval on anything metered or rate-limited.
- Make sure your check's load is negligible next to real traffic, not a meaningful fraction of it.

> The best interval is the slowest one you can tolerate for that endpoint's blast radius — fast where downtime hurts, relaxed everywhere else.

Pick deliberately, tier your endpoints, and let your thresholds do the noise filtering. Do that and you'll catch the outages that matter in seconds without drowning in false alarms.

[**Start monitoring free →**](/register)
