import type { Metadata } from "next";
import type { ReactNode } from "react";
import { JsonLd } from "@/components/JsonLd";
import type { PublicStatusPage } from "@/lib/api";

const siteUrl = process.env.NEXT_PUBLIC_SITE_URL ?? "https://pingdan.dev";
const apiUrl = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";

async function getStatusPage(slug: string): Promise<PublicStatusPage | null> {
  try {
    const response = await fetch(`${apiUrl}/public/status/${encodeURIComponent(slug)}`, {
      next: { revalidate: 300 },
    });
    if (!response.ok) return null;
    return response.json() as Promise<PublicStatusPage>;
  } catch {
    return null;
  }
}

export async function generateMetadata({ params }: { params: { slug: string } }): Promise<Metadata> {
  const page = await getStatusPage(params.slug);
  if (!page) {
    return {
      title: "Status Page",
      robots: { index: false, follow: true },
    };
  }

  const description = page.description || `Live availability and uptime information for ${page.title}.`;
  return {
    title: `${page.title} Service Status`,
    description,
    alternates: { canonical: `/status/${params.slug}` },
    openGraph: {
      type: "website",
      title: `${page.title} Service Status`,
      description,
      url: `/status/${params.slug}`,
    },
    robots: { index: true, follow: true },
  };
}

export default async function StatusLayout({ children, params }: { children: ReactNode; params: { slug: string } }) {
  const page = await getStatusPage(params.slug);
  if (!page) return children;

  const description = page.description || `Live availability and uptime information for ${page.title}.`;
  return (
    <>
      <JsonLd data={{
        "@context": "https://schema.org",
        "@type": "WebPage",
        name: `${page.title} Service Status`,
        description,
        url: `${siteUrl}/status/${params.slug}`,
        dateModified: page.updatedAt,
        isPartOf: { "@type": "WebSite", "@id": `${siteUrl}/#website`, name: "pingdan" },
      }} />
      {children}
    </>
  );
}
