import Link from "next/link";
import { MarketingNav } from "@/components/MarketingNav";
import { Footer } from "@/components/Footer";
import { Breadcrumbs, JsonLd } from "@/components/JsonLd";

const siteUrl = process.env.NEXT_PUBLIC_SITE_URL ?? "https://pingdan.dev";

export type Solution = {
  path: string;
  eyebrow: string;
  title: string;
  lede: string;
  introTitle: string;
  intro: string;
  benefits: { title: string; text: string }[];
  workflow: { title: string; text: string }[];
  faq: { question: string; answer: string }[];
  guides: { href: string; title: string; text: string }[];
};

export function SolutionPage({ solution }: { solution: Solution }) {
  const faqJsonLd = {
    "@context": "https://schema.org",
    "@type": "FAQPage",
    mainEntity: solution.faq.map((item) => ({
      "@type": "Question",
      name: item.question,
      acceptedAnswer: { "@type": "Answer", text: item.answer },
    })),
  };
  const serviceJsonLd = {
    "@context": "https://schema.org",
    "@type": "Service",
    name: solution.eyebrow,
    description: solution.lede,
    url: `${siteUrl}${solution.path}`,
    provider: { "@id": `${siteUrl}/#organization` },
    areaServed: "Worldwide",
    offers: { "@type": "Offer", price: "0", priceCurrency: "USD" },
  };

  return (
    <div className="mkt">
      <JsonLd data={faqJsonLd} />
      <JsonLd data={serviceJsonLd} />
      <Breadcrumbs trail={[{ name: solution.eyebrow, path: solution.path }]} />
      <MarketingNav />

      <section className="hero">
        <div className="mkt-wrap">
          <span className="eyebrow">{solution.eyebrow}</span>
          <h1>{solution.title}</h1>
          <p className="lede">{solution.lede}</p>
          <div className="cta-row">
            <Link href="/register" className="button-link primary btn-lg">Start monitoring free</Link>
            <Link href="/docs" className="button-link btn-lg">View setup docs</Link>
          </div>
          <div className="trust">No credit card · 1-minute checks · Every feature included</div>
        </div>
      </section>

      <section className="mkt-section tight">
        <div className="mkt-wrap solution-copy">
          <h2>{solution.introTitle}</h2>
          <p>{solution.intro}</p>
          <ul className="solution-checklist">
            {solution.benefits.map((benefit) => (
              <li key={benefit.title}><strong>{benefit.title}</strong>{benefit.text}</li>
            ))}
          </ul>
        </div>
      </section>

      <section className="mkt-section">
        <div className="mkt-wrap">
          <div className="section-head"><span className="eyebrow">How it works</span><h2>Useful signal in three steps</h2></div>
          <div className="feature-grid">
            {solution.workflow.map((step, index) => (
              <div className="feature" key={step.title}><div className="ico">{index + 1}</div><h3>{step.title}</h3><p>{step.text}</p></div>
            ))}
          </div>
        </div>
      </section>

      <section className="mkt-section tight">
        <div className="mkt-wrap">
          <div className="section-head"><span className="eyebrow">Learn more</span><h2>Practical monitoring guides</h2></div>
          <div className="blog-grid">
            {solution.guides.map((guide) => (
              <Link href={guide.href} className="blog-card" key={guide.href}><h3>{guide.title}</h3><p>{guide.text}</p><span className="eyebrow">Read guide →</span></Link>
            ))}
          </div>
        </div>
      </section>

      <section className="mkt-section">
        <div className="mkt-wrap solution-copy">
          <div className="section-head"><span className="eyebrow">FAQ</span><h2>Common questions</h2></div>
          <div className="stack" style={{ gap: "1rem" }}>
            {solution.faq.map((item) => <div className="card" key={item.question}><h3>{item.question}</h3><p className="muted" style={{ margin: 0 }}>{item.answer}</p></div>)}
          </div>
        </div>
      </section>

      <section className="mkt-section"><div className="mkt-wrap"><div className="cta-band"><h2>Know before your users do</h2><p>Create your first monitor in under a minute. Every feature is free.</p><Link href="/register" className="button-link primary btn-lg">Create a free monitor</Link></div></div></section>
      <Footer />
    </div>
  );
}
