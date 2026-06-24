import Link from "next/link";
import type { Metadata } from "next";
import { MarketingNav } from "@/components/MarketingNav";
import { Footer } from "@/components/Footer";

export const metadata: Metadata = {
  title: "Pricing — pingdan",
  description:
    "pingdan is free. Every feature included — unlimited monitors, 1-minute intervals, all alert channels, and full history. No credit card required.",
  alternates: { canonical: "/pricing" },
};

const plans = [
  {
    plan: "Free",
    amt: "$0",
    per: "/mo",
    desc: "Everything pingdan offers, free for everyone.",
    cta: "Get started free",
    featured: true,
    features: ["Unlimited monitors", "1-minute intervals", "All assertion types", "Email + Telegram alerts", "Full history", "Latency percentiles"],
  },
];

const faqs = [
  { q: "Is pingdan really free?", a: "Yes. Every feature is free for everyone — unlimited monitors, 1-minute intervals, all alert channels, and full history. No credit card required." },
  { q: "Are any features locked behind a paid plan?", a: "No. There are no paid plans. You get the full product with no limits or upsells." },
  { q: "What counts as a monitor?", a: "One HTTP endpoint you watch on a schedule. Each monitor can carry as many assertions as you like." },
  { q: "How do alerts work?", a: "When an endpoint fails its assertions for your configured threshold, we notify the channels attached to it — email and/or Telegram." },
];

export default function PricingPage() {
  return (
    <div className="mkt">
      <MarketingNav />

      <section className="hero" style={{ padding: "4rem 0 1rem" }}>
        <div className="mkt-wrap">
          <span className="eyebrow">Pricing</span>
          <h1 style={{ fontSize: "clamp(2rem,4vw,3rem)" }}>It&apos;s free. All of it.</h1>
          <p className="lede">Every feature, no limits, no credit card. pingdan is free for everyone.</p>
        </div>
      </section>

      <section className="mkt-section tight">
        <div className="mkt-wrap">
          <div className="price-grid">
            {plans.map((p) => (
              <div className={`price-card ${p.featured ? "featured" : ""}`} key={p.plan}>
                {p.featured && <span className="tag">Most popular</span>}
                <div className="plan">{p.plan}</div>
                <div className="amt">{p.amt}<small>{p.per}</small></div>
                <div className="desc">{p.desc}</div>
                <ul>
                  {p.features.map((f) => (
                    <li key={f}><span className="tick">✓</span><span>{f}</span></li>
                  ))}
                </ul>
                <Link href="/register">
                  <button className={p.featured ? "primary" : ""} style={{ width: "100%" }}>{p.cta}</button>
                </Link>
              </div>
            ))}
          </div>
        </div>
      </section>

      <section className="mkt-section">
        <div className="mkt-wrap" style={{ maxWidth: 760 }}>
          <div className="section-head">
            <span className="eyebrow">FAQ</span>
            <h2>Questions, answered</h2>
          </div>
          <div className="stack" style={{ gap: "1rem" }}>
            {faqs.map((f) => (
              <div className="card" key={f.q}>
                <h3 style={{ marginBottom: "0.4rem" }}>{f.q}</h3>
                <p className="muted" style={{ margin: 0 }}>{f.a}</p>
              </div>
            ))}
          </div>
        </div>
      </section>

      <Footer />
    </div>
  );
}
