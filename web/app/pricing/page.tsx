import Link from "next/link";
import type { Metadata } from "next";
import { MarketingNav } from "@/components/MarketingNav";
import { Footer } from "@/components/Footer";

export const metadata: Metadata = {
  title: "Pricing — pingdan",
  description:
    "Simple, transparent pricing for uptime monitoring. Start free, upgrade for shorter intervals, more monitors, and team features.",
  alternates: { canonical: "/pricing" },
};

const plans = [
  {
    plan: "Free",
    amt: "$0",
    per: "/mo",
    desc: "For side projects and trying things out.",
    cta: "Start free",
    featured: false,
    features: ["5 monitors", "5-minute minimum interval", "Status + body assertions", "Email alerts", "24-hour history"],
  },
  {
    plan: "Pro",
    amt: "$12",
    per: "/mo",
    desc: "For developers running production services.",
    cta: "Start 14-day trial",
    featured: true,
    features: ["50 monitors", "1-minute intervals", "All assertion types", "Email + Telegram alerts", "30-day history", "Latency percentiles"],
  },
  {
    plan: "Team",
    amt: "$39",
    per: "/mo",
    desc: "For teams who need shared visibility.",
    cta: "Contact sales",
    featured: false,
    features: ["200 monitors", "1-minute intervals", "Everything in Pro", "Multiple team members", "90-day history", "Priority support"],
  },
];

const faqs = [
  { q: "Is there really a free plan?", a: "Yes. The Free plan includes 5 monitors at a 5-minute interval, forever. No credit card required to start." },
  { q: "Can I change plans later?", a: "Anytime. Upgrades take effect immediately and downgrades apply at the end of your billing period." },
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
          <h1 style={{ fontSize: "clamp(2rem,4vw,3rem)" }}>Simple, honest pricing</h1>
          <p className="lede">Start free. Upgrade when you need faster checks and more monitors.</p>
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
