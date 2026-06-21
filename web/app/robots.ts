import type { MetadataRoute } from "next";

const siteUrl = process.env.NEXT_PUBLIC_SITE_URL ?? "https://pingdan.dev";

export default function robots(): MetadataRoute.Robots {
  return {
    rules: {
      userAgent: "*",
      allow: "/",
      // Keep authenticated app routes out of the index.
      disallow: ["/dashboard", "/endpoints", "/channels", "/auth/"],
    },
    sitemap: `${siteUrl}/sitemap.xml`,
  };
}
