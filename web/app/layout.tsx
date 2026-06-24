import type { ReactNode } from "react";
import type { Metadata, Viewport } from "next";
import { JsonLd } from "@/components/JsonLd";
import "./globals.css";

const siteUrl = process.env.NEXT_PUBLIC_SITE_URL ?? "https://pingdan.dev";

export const metadata: Metadata = {
  metadataBase: new URL(siteUrl),
  title: {
    default: "pingdan — Uptime & API monitoring with deep assertions",
    template: "%s | pingdan",
  },
  description:
    "Monitor HTTP endpoints with deep assertions on status, headers, body and JSON path. Response-time charts, uptime history, and instant email & Telegram alerts.",
  keywords: [
    "uptime monitoring",
    "API monitoring",
    "HTTP monitoring",
    "endpoint monitoring",
    "response time",
    "status page",
    "synthetic monitoring",
  ],
  openGraph: {
    type: "website",
    siteName: "pingdan",
    title: "pingdan — Uptime & API monitoring with deep assertions",
    description:
      "Monitor HTTP endpoints with deep assertions, response-time charts, and instant alerts.",
    url: siteUrl,
  },
  twitter: {
    card: "summary_large_image",
    title: "pingdan — Uptime & API monitoring",
    description:
      "Monitor HTTP endpoints with deep assertions, response-time charts, and instant alerts.",
  },
  alternates: { canonical: "/" },
  robots: {
    index: true,
    follow: true,
    googleBot: { index: true, follow: true, "max-image-preview": "large", "max-snippet": -1 },
  },
};

export const viewport: Viewport = {
  themeColor: "#0b1120",
};

const orgJsonLd = {
  "@context": "https://schema.org",
  "@type": "Organization",
  name: "pingdan",
  url: siteUrl,
  logo: `${siteUrl}/icon.png`,
  contactPoint: {
    "@type": "ContactPoint",
    email: "support@pingdan.dev",
    contactType: "customer support",
  },
};

const siteJsonLd = {
  "@context": "https://schema.org",
  "@type": "WebSite",
  name: "pingdan",
  url: siteUrl,
};

export default function RootLayout({ children }: { children: ReactNode }) {
  return (
    <html lang="en">
      <body>
        <JsonLd data={orgJsonLd} />
        <JsonLd data={siteJsonLd} />
        {children}
      </body>
    </html>
  );
}
