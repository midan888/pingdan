import Link from "next/link";
import type { Metadata } from "next";
import { notFound } from "next/navigation";
import { MarketingNav } from "@/components/MarketingNav";
import { Footer } from "@/components/Footer";
import { Breadcrumbs, JsonLd } from "@/components/JsonLd";
import { getPost, getAllSlugs, getRelatedPosts, formatDate } from "@/lib/blog";

const siteUrl = process.env.NEXT_PUBLIC_SITE_URL ?? "https://pingdan.dev";

type Params = { params: { slug: string } };

export function generateStaticParams() {
  return getAllSlugs().map((slug) => ({ slug }));
}

export function generateMetadata({ params }: Params): Metadata {
  const post = getPost(params.slug);
  if (!post) return { title: "Article not found" };
  const url = `${siteUrl}/blog/${post.slug}`;
  return {
    title: post.title,
    description: post.description,
    keywords: post.keywords,
    authors: [{ name: post.author }],
    alternates: { canonical: `/blog/${post.slug}` },
    openGraph: {
      type: "article",
      title: post.title,
      description: post.description,
      url,
      publishedTime: post.date,
      authors: [post.author],
      tags: post.tags,
    },
    twitter: {
      card: "summary_large_image",
      title: post.title,
      description: post.description,
    },
  };
}

export default function ArticlePage({ params }: Params) {
  const post = getPost(params.slug);
  if (!post) notFound();
  const related = getRelatedPosts(post.slug, 3);
  const url = `${siteUrl}/blog/${post.slug}`;

  const articleJsonLd = {
    "@context": "https://schema.org",
    "@type": "BlogPosting",
    mainEntityOfPage: { "@type": "WebPage", "@id": url },
    headline: post.title,
    description: post.description,
    datePublished: post.date,
    dateModified: post.date,
    author: { "@type": "Organization", name: post.author, url: siteUrl },
    publisher: {
      "@type": "Organization",
      name: "pingdan",
      logo: { "@type": "ImageObject", url: `${siteUrl}/icon.png` },
    },
    keywords: post.keywords.join(", "),
    image: `${url}/opengraph-image`,
    url,
  };

  return (
    <div className="mkt">
      <Breadcrumbs
        trail={[
          { name: "Blog", path: "/blog" },
          { name: post.title, path: `/blog/${post.slug}` },
        ]}
      />
      <JsonLd data={articleJsonLd} />
      <MarketingNav />

      <article className="mkt-section tight">
        <div className="mkt-wrap" style={{ maxWidth: 760 }}>
          <div className="article-head">
            <div className="blog-card-tags">
              {post.tags.map((t) => (
                <span className="tag" key={t}>
                  {t}
                </span>
              ))}
            </div>
            <h1>{post.title}</h1>
            <p className="lede">{post.description}</p>
            <div className="blog-meta">
              <span>By {post.author}</span>
              <span>·</span>
              <span>{formatDate(post.date)}</span>
              <span>·</span>
              <span>{post.readingMinutes} min read</span>
            </div>
          </div>

          <div className="prose" dangerouslySetInnerHTML={{ __html: post.html }} />

          <div className="article-cta">
            <h3>Monitor it before it breaks</h3>
            <p className="muted">
              pingdan checks your endpoints around the clock with deep assertions and
              sends instant alerts through the channels your team watches. Free to start.
            </p>
            <Link href="/register" className="button-link primary btn-lg">Start monitoring free</Link>
          </div>
        </div>
      </article>

      {related.length > 0 && (
        <section className="mkt-section tight">
          <div className="mkt-wrap">
            <div className="section-head">
              <span className="eyebrow">Keep reading</span>
              <h2>Related articles</h2>
            </div>
            <div className="blog-grid">
              {related.map((p) => (
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
          </div>
        </section>
      )}

      <Footer />
    </div>
  );
}
