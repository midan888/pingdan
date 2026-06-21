import Link from "next/link";
import type { Metadata } from "next";
import { MarketingNav } from "@/components/MarketingNav";
import { Footer } from "@/components/Footer";
import { MiniDashboardPreview } from "@/components/MiniDashboardPreview";

export const metadata: Metadata = {
  title: "pingdan — Uptime & API monitoring with deep assertions",
  description:
    "Monitor HTTP endpoints from 1-minute intervals. Assert on status codes, headers, body and JSON path, watch response-time charts, and get instant alerts via email and Telegram.",
  alternates: { canonical: "/" },
};

const features = [
  { icon: "✓", title: "Deep assertions", desc: "Go beyond status codes. Validate headers, response body, JSON paths and response time on every check." },
  { icon: "⌁", title: "Response-time charts", desc: "Per-endpoint bar charts, p50/p95 latency, and uptime history so you spot regressions early." },
  { icon: "◷", title: "1-minute checks", desc: "Fixed, predictable intervals from 1 to 8 minutes. Know about incidents fast." },
  { icon: "✉", title: "Instant alerts", desc: "Email and Telegram notifications the moment an endpoint goes down — and again when it recovers." },
  { icon: "⛒", title: "Failure thresholds", desc: "Avoid noise. Only alert after N consecutive failures, with full per-check failure detail." },
  { icon: "⚡", title: "Set up in a minute", desc: "Paste a URL, pick what makes a check pass, and you're monitoring. No agents, no config files." },
];

const stats = [
  { n: "1 min", l: "Fastest check interval" },
  { n: "5", l: "Assertion sources" },
  { n: "p95", l: "Latency tracking" },
  { n: "2", l: "Alert channels" },
];

export default function Landing() {
  return (
    <div className="mkt">
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
            <Link href="/register"><button className="primary btn-lg">Start monitoring free</button></Link>
            <Link href="/docs"><button className="btn-lg">Read the docs</button></Link>
          </div>
          <div className="trust">No credit card · Free tier · Cancel anytime</div>
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

      {/* CTA band */}
      <section className="mkt-section">
        <div className="mkt-wrap">
          <div className="cta-band">
            <h2>Start monitoring in under a minute</h2>
            <p>Paste a URL, choose what makes a check pass, and you&apos;re live. Free to start.</p>
            <Link href="/register"><button className="primary btn-lg">Create your free account</button></Link>
          </div>
        </div>
      </section>

      <Footer />
    </div>
  );
}
