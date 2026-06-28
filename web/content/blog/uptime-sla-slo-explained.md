---
title: "SLA, SLO and SLI Explained: The Uptime Math Behind Nines"
description: "SLA SLO SLI uptime, demystified for engineers. Learn how availability nines map to real downtime, how error budgets work, and how to measure it correctly."
date: 2026-06-22
author: pingdan
tags: [SLA, Reliability]
keywords: [sla slo sli, uptime sla, what is an slo, error budget, nines of availability, 99.9 uptime, availability percentage]
---

Someone in your channel just typed "we're 99.9% available, right?" and nobody could answer with certainty. That single percentage hides a surprising amount of math, a set of contractual promises, and a measurement problem that trips up even seasoned teams. Before you commit to a number on a status page or a contract, it pays to know exactly what you're promising and how you'd prove it.

The three terms that get tangled together are SLI, SLO and SLA. They nest cleanly once you see them in the right order.

## SLI, SLO, SLA: three layers, not three synonyms

| Term | What it is | Example |
| --- | --- | --- |
| **SLI** (Service Level Indicator) | A *measurement* of one aspect of service quality | Successful HTTP responses / total responses |
| **SLO** (Service Level Objective) | A *target* you set for that indicator | 99.9% successful over 30 days |
| **SLA** (Service Level Agreement) | A *contract* with consequences if the SLO is missed | Credit 10% of the bill if availability drops below 99.5% |

The dependency runs one direction. You can't define an SLO without an SLI to measure against, and you shouldn't sign an SLA without an SLO you're confident you can hit. Most teams live at the SLO level internally; the SLA is the money-on-the-line version, almost always set looser than your internal SLO so you keep headroom.

## How availability is actually calculated

Availability is just good time over total time:

```
availability = good_minutes / total_minutes
```

The hard part isn't the division, it's defining `good`. Is a request "good" if it returns *any* response? Within 500ms? If the JSON body actually contains the field your client expects? Each definition produces a different number from the same traffic. Pin down the SLI before you argue about the SLO.

## The nines table

"Nines" is shorthand for the number of nines in your availability percentage. Here's what each tier costs you in allowable downtime:

| Availability | Per year | Per month | Per week |
| --- | --- | --- | --- |
| 99% ("two nines") | 3.65 days | 7.31 hours | 1.68 hours |
| 99.9% ("three nines") | 8.77 hours | 43.8 minutes | 10.1 minutes |
| 99.99% ("four nines") | 52.6 minutes | 4.38 minutes | 1.01 minutes |
| 99.999% ("five nines") | 5.26 minutes | 26.3 seconds | 6.05 seconds |

The jump from 99.9% to 99.99% means going from "I can lose most of a working day" to "I have under an hour for the entire year, including deploys, cert renewals and that 3am page." Each extra nine roughly multiplies your operational cost. Most products don't need five nines, and claiming them without the infrastructure to back it up is how SLAs turn into refunds.

## Error budgets: the point of all this

If your SLO is 99.9%, your **error budget** is the inverse: 0.1%, or about 43.8 minutes of allowed downtime per month. That budget isn't a failure to be ashamed of, it's a resource to spend.

This reframes reliability as a tradeoff instead of an absolute. Under budget, you ship aggressively and run risky migrations. Burn through it, and you freeze feature work to put the budget back into stability. The error budget turns "should we deploy on Friday?" from a vibe into a number.

A useful way to think about it:

- **Budget remaining** → ship features, take risks.
- **Budget exhausted** → stop shipping, fix reliability.
- **Burning budget fast** → page someone now, you have a live incident.

## A 200 OK can still be lying to you

Here's the trap. The easiest SLI to collect is "did the server return 2xx?" — but that measurement inflates your numbers. An endpoint can return `200 OK` while serving a stale cache, an empty array where there should be data, or a JSON body missing the field every client depends on. Your dashboard stays green. Your users are stuck. Your SLI says you're fine.

Measuring availability honestly requires checking the *content* of the response, not just the status line. A check that asserts the body contains the expected fields, matches a schema, or stays under a latency threshold gives you an SLI that reflects reality — the kind of [deep assertion that separates real API monitoring from a naive ping](/blog/api-monitoring-best-practices), and the difference between knowing your API is *up* and knowing it's *working*.

## Why you measure from outside

There's a second reason your internal numbers lie: they're measured from inside the building. Compute availability from your own application logs and you miss every failure that happens before the request reaches your code — DNS issues, expired TLS certificates, load balancer misconfigurations, or your whole service being unreachable. Your servers can't log a request they never received, so the outage looks like a quiet, healthy period in your metrics.

External monitoring closes that gap by probing your service the way a user would, from outside your network. That's the foundation of an honest SLI; [what is uptime monitoring](/blog/what-is-uptime-monitoring) walks through why the vantage point matters.

## Tracking your numbers in practice

To actually run SLOs instead of just talking about them, you need four things:

1. **A clear SLI definition** — what counts as a good request, including content and latency, not just status code.
2. **An external probe** at a sensible interval. Check too rarely and a short outage hides between samples; the right [monitoring frequency](/blog/how-often-should-you-monitor) depends on the tightness of your SLO.
3. **Uptime history** to compute availability over rolling 30-day windows and see whether you're on track.
4. **Alerts** — email and Telegram — that fire on budget burn, not just hard-down, so you catch slow degradations before they eat the month.

Response-time charts and uptime history turn all of this from a quarterly spreadsheet exercise into something you glance at any day.

> A nine you can't measure is just marketing. The math is the easy part — the discipline is in measuring `good` honestly, from the outside, every single minute.

[**Start monitoring free →**](/register)
