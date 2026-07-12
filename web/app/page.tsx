import Link from "next/link";
import type { Metadata } from "next";
import { MarketingNav } from "@/components/MarketingNav";
import { Footer } from "@/components/Footer";
import { MiniDashboardPreview } from "@/components/MiniDashboardPreview";
import { JsonLd } from "@/components/JsonLd";

const siteUrl = process.env.NEXT_PUBLIC_SITE_URL ?? "https://pingdan.dev";

export const metadata: Metadata = {
  title: { absolute: "pingdan — Uptime & API monitoring with deep assertions" },
  description:
    "Monitor HTTP endpoints from 1-minute intervals. Assert on status codes, headers, body and JSON path, watch response-time charts, and get instant alerts across email, chat, paging, webhooks and SMS.",
  alternates: { canonical: "/" },
};

const features = [
  { icon: "✓", title: "Deep assertions", desc: "Go beyond status codes. Validate headers, response body, JSON paths and response time on every check." },
  { icon: "⌁", title: "Response-time charts", desc: "Per-endpoint bar charts, p50/p95 latency, and uptime history so you spot regressions early." },
  { icon: "◷", title: "1-minute checks", desc: "Custom intervals from 1 minute to 7 days. Know about incidents fast." },
  { icon: "✉", title: "Instant alerts", desc: "Route alerts through email, chat, webhooks, PagerDuty, ntfy, Pushover, SMS and Opsgenie." },
  { icon: "⛒", title: "Failure thresholds", desc: "Avoid noise. Only alert after N consecutive failures, with full per-check failure detail." },
  { icon: "⚡", title: "Set up in a minute", desc: "Paste a URL, pick what makes a check pass, and you're monitoring. No agents, no config files." },
];

const stats = [
  { n: "1 min", l: "Fastest check interval" },
  { n: "5", l: "Assertion sources" },
  { n: "p95", l: "Latency tracking" },
  { n: "11", l: "Alert channels" },
];

const appJsonLd = {
  "@context": "https://schema.org",
  "@type": "SoftwareApplication",
  name: "pingdan",
  applicationCategory: "DeveloperApplication",
  operatingSystem: "Web",
  url: siteUrl,
  description:
    "Uptime & API monitoring with deep assertions on status, headers, body and JSON path. Response-time charts, uptime history, and instant multi-channel alerts.",
  featureList: [
    "1-minute uptime checks",
    "HTTP status, header, body, JSON path and latency assertions",
    "Response-time charts and latency percentiles",
    "SSL certificate expiry monitoring",
    "Public status pages",
    "Multi-channel downtime and recovery alerts",
  ],
  offers: {
    "@type": "Offer",
    price: "0",
    priceCurrency: "USD",
  },
};

export default function Landing() {
  return (
    <div className="mkt">
      <JsonLd data={appJsonLd} />
      <MarketingNav />

      {/* hero */}
      <section className="hero">
        <div className="mkt-wrap">
          <span className="badge-soft"><span className="dot up" /> Monitoring that actually asserts</span>
          <h1>Know the moment your API breaks.</h1>
          <p className="lede">
            pingdan watches your HTTP endpoints, asserts on everything that matters, and alerts you
            before your customers notice — with charts that show exactly what happened.
          </p>
          <div className="cta-row">
            <Link href="/register" className="button-link primary btn-lg">Start monitoring free</Link>
            <Link href="/docs" className="button-link btn-lg">Read the docs</Link>
          </div>
          <div className="trust">No credit card · Free forever · Every feature included</div>
          <MiniDashboardPreview />
        </div>
      </section>

      {/* stats */}
      <section className="mkt-section tight">
        <div className="mkt-wrap">
          <div className="stat-band">
            {stats.map((s) => (
              <div key={s.l}>
                <div className="n">{s.n}</div>
                <div className="l">{s.l}</div>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* features */}
      <section className="mkt-section" id="features">
        <div className="mkt-wrap">
          <div className="section-head">
            <span className="eyebrow">Why pingdan</span>
            <h2>Everything you need to trust your uptime</h2>
            <p>Monitoring built for engineers who want signal, not noise.</p>
          </div>
          <div className="feature-grid">
            {features.map((f) => (
              <div className="feature" key={f.title}>
                <div className="ico">{f.icon}</div>
                <h3>{f.title}</h3>
                <p>{f.desc}</p>
              </div>
            ))}
          </div>
        </div>
      </section>

      <section className="mkt-section tight">
        <div className="mkt-wrap">
          <div className="section-head">
            <span className="eyebrow">Monitoring by use case</span>
            <h2>Protect every public endpoint</h2>
            <p>Start with availability, then validate the response and the certificate behind it.</p>
          </div>
          <div className="feature-grid">
            <Link href="/uptime-monitoring" className="feature solution-link">
              <h3>Website uptime monitoring</h3>
              <p>Check availability every minute, measure response time, and alert only after the failure threshold you choose.</p>
              <span>Explore uptime monitoring →</span>
            </Link>
            <Link href="/api-monitoring" className="feature solution-link">
              <h3>API monitoring</h3>
              <p>Validate status codes, headers, response bodies, JSON paths, and latency—not merely whether a server answered.</p>
              <span>Explore API monitoring →</span>
            </Link>
            <Link href="/ssl-certificate-monitoring" className="feature solution-link">
              <h3>SSL certificate monitoring</h3>
              <p>Track HTTPS certificate expiry automatically and send warnings before a forgotten renewal becomes an outage.</p>
              <span>Explore SSL monitoring →</span>
            </Link>
          </div>
        </div>
      </section>

      {/* CTA band */}
      <section className="mkt-section">
        <div className="mkt-wrap">
          <div className="cta-band">
            <h2>Start monitoring in under a minute</h2>
            <p>Paste a URL, choose what makes a check pass, and you&apos;re live. Free to start.</p>
            <Link href="/register" className="button-link primary btn-lg">Create your free account</Link>
          </div>
        </div>
      </section>

      <Footer />
    </div>
  );
}
