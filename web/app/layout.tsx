import type { ReactNode } from "react";
import type { Metadata, Viewport } from "next";
import { JsonLd } from "@/components/JsonLd";
import "./globals.css";

const siteUrl = process.env.NEXT_PUBLIC_SITE_URL ?? "https://pingdan.dev";

export const metadata: Metadata = {
  applicationName: "pingdan",
  category: "technology",
  creator: "pingdan",
  publisher: "pingdan",
  referrer: "origin-when-cross-origin",
  metadataBase: new URL(siteUrl),
  title: {
    default: "pingdan — Uptime & API monitoring with deep assertions",
    template: "%s | pingdan",
  },
  description:
    "Monitor HTTP endpoints with deep assertions on status, headers, body and JSON path. Response-time charts, uptime history, and instant alerts across email, chat, webhooks, paging and SMS.",
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
  icons: {
    icon: [{ url: "/icon.svg", type: "image/svg+xml" }, { url: "/favicon.ico" }],
    apple: "/apple-icon.png",
  },
  manifest: "/manifest.webmanifest",
  alternates: { canonical: "/" },
  robots: {
    index: true,
    follow: true,
    googleBot: { index: true, follow: true, "max-image-preview": "large", "max-snippet": -1 },
  },
};

export const viewport: Viewport = {
  themeColor: "#0b1120",
  // viewport-fit=cover lets the bottom tab bar extend under the iPhone home
  // indicator using env(safe-area-inset-bottom)
  viewportFit: "cover",
};

const orgJsonLd = {
  "@context": "https://schema.org",
  "@type": "Organization",
  "@id": `${siteUrl}/#organization`,
  name: "pingdan",
  url: siteUrl,
  logo: `${siteUrl}/icon.png`,
  email: "support@pingdan.dev",
  contactPoint: {
    "@type": "ContactPoint",
    email: "support@pingdan.dev",
    contactType: "customer support",
  },
};

const siteJsonLd = {
  "@context": "https://schema.org",
  "@type": "WebSite",
  "@id": `${siteUrl}/#website`,
  name: "pingdan",
  alternateName: "pingdan uptime monitoring",
  url: siteUrl,
  publisher: { "@id": `${siteUrl}/#organization` },
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
