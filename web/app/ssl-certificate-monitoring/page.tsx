import type { Metadata } from "next";
import { SolutionPage, type Solution } from "@/components/SolutionPage";

const description = "Free SSL certificate monitoring for HTTPS endpoints with automatic expiry checks, certificate health visibility, and alerts before renewal failures cause outages.";

export const metadata: Metadata = {
  title: "Free SSL Certificate Expiry Monitoring & Alerts",
  description,
  keywords: ["SSL certificate monitoring", "SSL expiry monitor", "TLS certificate expiry alert", "certificate monitoring", "HTTPS monitoring"],
  alternates: { canonical: "/ssl-certificate-monitoring" },
  openGraph: { title: "SSL Certificate Expiry Monitoring | pingdan", description, url: "/ssl-certificate-monitoring" },
  twitter: { card: "summary_large_image", title: "SSL Certificate Expiry Monitoring | pingdan", description },
};

const solution: Solution = {
  path: "/ssl-certificate-monitoring",
  eyebrow: "SSL certificate monitoring",
  title: "Renew every TLS certificate before it becomes an outage",
  lede: "Automatically inspect certificates for your HTTPS monitors, see expiry and handshake errors, and alert your team when renewal needs attention.",
  introTitle: "Automation still needs an independent expiry check",
  intro: "Certificate renewal can fail because of DNS changes, invalid challenge routes, expired credentials, rate limits, or a broken deployment. pingdan checks HTTPS endpoints independently of the renewal system, records the leaf certificate expiry, and warns attached channels once the certificate reaches the alert window.",
  benefits: [
    { title: "Automatic HTTPS coverage", text: "Every enabled HTTPS endpoint is included in the certificate sweep without a second monitor to configure." },
    { title: "Expiry visibility", text: "See the exact expiry date, last check time, and remaining lifetime from the endpoint view." },
    { title: "Handshake diagnostics", text: "Surface TLS connection and certificate inspection errors instead of silently losing expiry data." },
    { title: "Multi-channel warnings", text: "Send expiry events to the same email, chat, webhook, paging, push, SMS, and incident tools used for uptime." },
  ],
  workflow: [
    { title: "Add an HTTPS URL", text: "Create a normal endpoint monitor. Certificate monitoring is enabled automatically for HTTPS." },
    { title: "Inspect the certificate", text: "pingdan performs a regular TLS handshake and stores the presented leaf certificate expiry." },
    { title: "Renew before impact", text: "When days remaining enter the warning window, attached channels receive the endpoint and expiry details." },
  ],
  faq: [
    { question: "Why monitor SSL certificate expiry if renewal is automatic?", answer: "Automatic renewal is a process that can fail. Independent monitoring verifies the certificate users actually receive and gives you time to fix DNS, challenge, credential, or deployment problems." },
    { question: "Which certificates does pingdan monitor?", answer: "pingdan monitors the leaf TLS certificate presented by each enabled HTTPS endpoint. HTTP-only endpoints do not have a TLS certificate to inspect." },
    { question: "How often are certificates checked?", answer: "Certificate sweeps run roughly twice daily, and a check can also be triggered on demand from the endpoint view. That cadence is appropriate because certificate expiry changes infrequently." },
    { question: "When does pingdan send an expiry warning?", answer: "Warnings begin when a certificate has 15 days or fewer remaining. The alert includes the endpoint, URL, days left, and exact expiry time." },
  ],
  guides: [
    { href: "/blog/monitor-ssl-certificate-expiry", title: "Prevent expired certificate outages", text: "Learn why automated renewal fails and how much warning time an operations team needs." },
    { href: "/uptime-monitoring", title: "Website uptime monitoring", text: "Pair certificate checks with one-minute availability, latency history, and downtime alerts." },
    { href: "/blog/website-down-alerts-email-telegram", title: "Route actionable alerts", text: "Choose alert channels people will see and include enough context to act immediately." },
  ],
};

export default function SslMonitoringPage() { return <SolutionPage solution={solution} />; }
