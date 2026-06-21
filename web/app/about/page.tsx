import Link from "next/link";
import type { Metadata } from "next";
import { MarketingNav } from "@/components/MarketingNav";
import { Footer } from "@/components/Footer";

export const metadata: Metadata = {
  title: "About & Contact — pingdan",
  description:
    "pingdan is uptime and API monitoring built for engineers. Learn what we believe and how to get in touch.",
  alternates: { canonical: "/about" },
};

const values = [
  { h: "Signal over noise", p: "Alerts should mean something. Thresholds, recovery notices and clear failure detail keep your inbox sane." },
  { h: "Depth by default", p: "A 200 isn't always healthy. We make asserting on the real response the easy path, not an add-on." },
  { h: "Fast to set up", p: "Monitoring you put off isn't monitoring. From URL to live in under a minute." },
];

export default function AboutPage() {
  return (
    <div className="mkt">
      <MarketingNav />

      <section className="hero" style={{ padding: "4rem 0 1rem" }}>
        <div className="mkt-wrap">
          <span className="eyebrow">About</span>
          <h1 style={{ fontSize: "clamp(2rem,4vw,3rem)" }}>Monitoring built by people who got paged at 3am</h1>
          <p className="lede">
            We built pingdan because every outage we lived through had the same root cause: a check that
            passed when it shouldn&apos;t have. So we made assertions the heart of the product.
          </p>
        </div>
      </section>

      <section className="mkt-section">
        <div className="mkt-wrap">
          <div className="section-head">
            <span className="eyebrow">What we believe</span>
            <h2>Principles we ship by</h2>
          </div>
          <div className="feature-grid">
            {values.map((v) => (
              <div className="feature" key={v.h}>
                <h3>{v.h}</h3>
                <p>{v.p}</p>
              </div>
            ))}
          </div>
        </div>
      </section>

      <section className="mkt-section tight" id="contact">
        <div className="mkt-wrap" style={{ maxWidth: 760 }}>
          <div className="section-head">
            <span className="eyebrow">Contact</span>
            <h2>Get in touch</h2>
            <p>Questions, feedback or a bug to report? We read everything.</p>
          </div>
          <div className="grid grid-2">
            <div className="contact-card">
              <h3>Support</h3>
              <p className="muted">For help with your account or monitors.</p>
              <p><a href="mailto:support@pingdan.dev">support@pingdan.dev</a></p>
            </div>
            <div className="contact-card">
              <h3>Sales</h3>
              <p className="muted">For Team plans and volume pricing.</p>
              <p><a href="mailto:sales@pingdan.dev">sales@pingdan.dev</a></p>
            </div>
          </div>
        </div>
      </section>

      <section className="mkt-section">
        <div className="mkt-wrap">
          <div className="cta-band">
            <h2>Ready to stop guessing?</h2>
            <p>Set up your first monitor free and see the difference depth makes.</p>
            <Link href="/register"><button className="primary btn-lg">Start free</button></Link>
          </div>
        </div>
      </section>

      <Footer />
    </div>
  );
}
