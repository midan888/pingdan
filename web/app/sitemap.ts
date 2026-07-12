import type { MetadataRoute } from "next";
import { getAllPosts } from "@/lib/blog";

const siteUrl = process.env.NEXT_PUBLIC_SITE_URL ?? "https://pingdan.dev";

export default function sitemap(): MetadataRoute.Sitemap {
  const routes = [
    { path: "/", priority: 1.0, changeFrequency: "weekly" as const },
    { path: "/features", priority: 0.8, changeFrequency: "monthly" as const },
    { path: "/uptime-monitoring", priority: 0.9, changeFrequency: "monthly" as const },
    { path: "/api-monitoring", priority: 0.9, changeFrequency: "monthly" as const },
    { path: "/ssl-certificate-monitoring", priority: 0.8, changeFrequency: "monthly" as const },
    { path: "/pricing", priority: 0.9, changeFrequency: "monthly" as const },
    { path: "/docs", priority: 0.7, changeFrequency: "monthly" as const },
    { path: "/blog", priority: 0.8, changeFrequency: "weekly" as const },
    { path: "/about", priority: 0.5, changeFrequency: "yearly" as const },
    // /login is intentionally omitted — it is noindex (no search value).
    { path: "/register", priority: 0.6, changeFrequency: "yearly" as const },
  ];

  const staticEntries: MetadataRoute.Sitemap = routes.map((r) => ({
    url: `${siteUrl}${r.path}`,
    changeFrequency: r.changeFrequency,
    priority: r.priority,
  }));

  const postEntries: MetadataRoute.Sitemap = getAllPosts().map((p) => ({
    url: `${siteUrl}/blog/${p.slug}`,
    lastModified: new Date(`${p.date}T00:00:00Z`),
    changeFrequency: "monthly" as const,
    priority: 0.7,
  }));

  return [...staticEntries, ...postEntries];
}
