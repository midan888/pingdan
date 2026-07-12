import Link from "next/link";
import type { Metadata } from "next";
import { MarketingNav } from "@/components/MarketingNav";
import { Footer } from "@/components/Footer";
import { Breadcrumbs, JsonLd } from "@/components/JsonLd";
import { getAllPosts, formatDate } from "@/lib/blog";

const siteUrl = process.env.NEXT_PUBLIC_SITE_URL ?? "https://pingdan.dev";

export const metadata: Metadata = {
  title: "Blog — Uptime & API monitoring guides",
  description:
    "Practical guides on uptime monitoring, API health checks, alerting, status codes, JSON assertions and SLAs — from the team behind pingdan.",
  keywords: [
    "uptime monitoring blog",
    "API monitoring guides",
    "website monitoring",
    "health checks",
    "alerting best practices",
  ],
  alternates: {
    canonical: "/blog",
    types: { "application/rss+xml": `${siteUrl}/feed.xml` },
  },
  openGraph: {
    type: "website",
    title: "pingdan Blog — Uptime & API monitoring guides",
    description:
      "Practical guides on uptime monitoring, API health checks, alerting and SLAs.",
    url: `${siteUrl}/blog`,
  },
};

export default function BlogIndexPage() {
  const posts = getAllPosts();
  const [featured, ...rest] = posts;

  const blogJsonLd = {
    "@context": "https://schema.org",
    "@type": "Blog",
    name: "pingdan Blog",
    description:
      "Practical guides on uptime monitoring, API health checks, alerting and SLAs.",
    url: `${siteUrl}/blog`,
    blogPost: posts.map((p) => ({
      "@type": "BlogPosting",
      headline: p.title,
      description: p.description,
      datePublished: p.date,
      url: `${siteUrl}/blog/${p.slug}`,
      author: { "@type": "Organization", name: p.author },
    })),
  };

  return (
    <div className="mkt">
      <Breadcrumbs trail={[{ name: "Blog", path: "/blog" }]} />
      <JsonLd data={blogJsonLd} />
      <MarketingNav />

      <section className="hero" style={{ padding: "4rem 0 1rem" }}>
        <div className="mkt-wrap">
          <span className="eyebrow">Blog</span>
          <h1 style={{ fontSize: "clamp(2rem,4vw,3rem)" }}>
            Monitoring, alerting & reliability — done right
          </h1>
          <p className="lede">
            Field-tested guides on uptime monitoring, API health checks, status codes,
            JSON assertions, SLAs and on-call alerting. Written by engineers who got
            paged at 3am so you don&apos;t have to.
          </p>
        </div>
      </section>

      <section className="mkt-section tight">
        <div className="mkt-wrap">
          {posts.length === 0 ? (
            <p className="muted">No articles yet — check back soon.</p>
          ) : (
            <>
              {featured && (
                <Link href={`/blog/${featured.slug}`} className="blog-featured">
                  <div className="blog-featured-body">
                    <span className="eyebrow">{featured.tags[0] ?? "Featured"}</span>
                    <h2>{featured.title}</h2>
                    <p>{featured.description}</p>
                    <div className="blog-meta">
                      <span>{formatDate(featured.date)}</span>
                      <span>·</span>
                      <span>{featured.readingMinutes} min read</span>
                    </div>
                  </div>
                </Link>
              )}

              <div className="blog-grid">
                {rest.map((p) => (
                  <Link key={p.slug} href={`/blog/${p.slug}`} className="blog-card">
                    <div className="blog-card-tags">
                      {p.tags.slice(0, 2).map((t) => (
                        <span className="tag" key={t}>
                          {t}
                        </span>
                      ))}
                    </div>
                    <h3>{p.title}</h3>
                    <p>{p.description}</p>
                    <div className="blog-meta">
                      <span>{formatDate(p.date)}</span>
                      <span>·</span>
                      <span>{p.readingMinutes} min read</span>
                    </div>
                  </Link>
                ))}
              </div>
            </>
          )}
        </div>
      </section>

      <section className="mkt-section">
        <div className="mkt-wrap">
          <div className="cta-band">
            <h2>Stop reading about downtime. Catch it.</h2>
            <p>Set up your first monitor free — assertions, charts and instant alerts.</p>
            <Link href="/register" className="button-link primary btn-lg">Start free</Link>
          </div>
        </div>
      </section>

      <Footer />
    </div>
  );
}
