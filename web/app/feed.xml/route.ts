import { getAllPosts } from "@/lib/blog";

const siteUrl = process.env.NEXT_PUBLIC_SITE_URL ?? "https://pingdan.dev";

function xml(value: string): string {
  return value.replace(/[<>&'\"]/g, (char) => ({
    "<": "&lt;", ">": "&gt;", "&": "&amp;", "'": "&apos;", '\"': "&quot;",
  })[char] ?? char);
}

export function GET() {
  const posts = getAllPosts();
  const items = posts.map((post) => `
    <item>
      <title>${xml(post.title)}</title>
      <link>${siteUrl}/blog/${post.slug}</link>
      <guid isPermaLink="true">${siteUrl}/blog/${post.slug}</guid>
      <description>${xml(post.description)}</description>
      <pubDate>${new Date(`${post.date}T00:00:00Z`).toUTCString()}</pubDate>
    </item>`).join("");

  return new Response(`<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0"><channel>
  <title>pingdan — Uptime &amp; API monitoring guides</title>
  <link>${siteUrl}/blog</link>
  <description>Practical guides on uptime monitoring, API health checks, alerting, and reliability.</description>
  <language>en</language>${items}
</channel></rss>`, {
    headers: { "Content-Type": "application/rss+xml; charset=utf-8", "Cache-Control": "public, max-age=3600" },
  });
}
