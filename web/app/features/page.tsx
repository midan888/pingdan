import Link from "next/link";
import type { Metadata } from "next";
import { MarketingNav } from "@/components/MarketingNav";
import { Footer } from "@/components/Footer";
import { Breadcrumbs } from "@/components/JsonLd";

export const metadata: Metadata = {
  title: "Uptime & API Monitoring Features",
  description:
    "Deep assertions on status, headers, body and JSON path. Response-time charts, uptime history, failure thresholds, and alert integrations for email, chat, webhooks, paging, mobile push, SMS and incident tools.",
  alternates: { canonical: "/features" },
};

const groups = [
  {
    eyebrow: "Validation",
    title: "Assert on everything, not just 200 OK",
    items: [
      { h: "Status code", p: "Require an exact code, a range, or anything-but with equals / not-equals / less-than / greater-than." },
      { h: "Response time", p: "Fail a check when latency crosses your threshold — catch slow degradation before a hard outage." },
      { h: "Headers", p: "Match any response header by name with contains, equals, regex and more." },
      { h: "Response body", p: "Assert the body contains a string, equals a value, or matches a regular expression." },
      { h: "JSON path", p: "Drill into JSON with dotted paths like data.status or items.0.id and compare the resolved value." },
    ],
  },
  {
    eyebrow: "Visibility",
    title: "Charts that explain what happened",
    items: [
      { h: "Response-time bars", p: "One bar per check, coloured by pass/fail, with hover detail for status, latency and errors." },
      { h: "Uptime timeline", p: "A compact status strip across your selected window — 1h, 6h, 24h or 7 days." },
      { h: "Latency percentiles", p: "Average, p50, p95, min and max so you understand the whole distribution, not just the mean." },
      { h: "Failure detail", p: "Every failed check records exactly which assertion failed and the actual value it saw." },
    ],
  },
  {
    eyebrow: "Operations",
    title: "Alerting that respects your attention",
    items: [
      { h: "Custom intervals", p: "Check every minute, every 6 hours, or once a day — any interval from 1 minute to 7 days." },
      { h: "Failure thresholds", p: "Only go down after N consecutive failures to filter out transient blips." },
      { h: "11 alert channels", p: "Route alerts to email, Telegram, Slack, Discord, Teams, webhooks, PagerDuty, ntfy, Pushover, Twilio SMS or Opsgenie, per endpoint." },
      { h: "Recovery notices", p: "Get told when things come back up, not just when they break." },
    ],
  },
];

export default function FeaturesPage() {
  return (
    <div className="mkt">
      <Breadcrumbs trail={[{ name: "Features", path: "/features" }]} />
      <MarketingNav />

      <section className="hero" style={{ padding: "4rem 0 1rem" }}>
        <div className="mkt-wrap">
          <span className="eyebrow">Features</span>
          <h1 style={{ fontSize: "clamp(2rem,4vw,3rem)" }}>Monitoring with depth</h1>
          <p className="lede">Everything pingdan checks, charts and alerts on — in one place.</p>
        </div>
      </section>

      {groups.map((g) => (
        <section className="mkt-section" key={g.title}>
          <div className="mkt-wrap">
            <div className="section-head">
              <span className="eyebrow">{g.eyebrow}</span>
              <h2>{g.title}</h2>
            </div>
            <div className="feature-grid">
              {g.items.map((it) => (
                <div className="feature" key={it.h}>
                  <h3>{it.h}</h3>
                  <p>{it.p}</p>
                </div>
              ))}
            </div>
          </div>
        </section>
      ))}

      <section className="mkt-section">
        <div className="mkt-wrap">
          <div className="cta-band">
            <h2>See it on your own endpoints</h2>
            <p>Spin up your first monitor in under a minute.</p>
            <Link href="/register" className="button-link primary btn-lg">Get started free</Link>
          </div>
        </div>
      </section>

      <Footer />
    </div>
  );
}
