import type { Metadata } from "next";
import { SolutionPage, type Solution } from "@/components/SolutionPage";

const description = "Free website uptime monitoring with 1-minute checks, response-time history, configurable failure thresholds, public status pages, and instant alerts.";

export const metadata: Metadata = {
  title: "Free Website Uptime Monitoring — 1-Minute Checks",
  description,
  keywords: ["uptime monitoring", "website uptime monitor", "free uptime monitoring", "website monitoring", "downtime alerts"],
  alternates: { canonical: "/uptime-monitoring" },
  openGraph: { title: "Free Website Uptime Monitoring | pingdan", description, url: "/uptime-monitoring" },
  twitter: { card: "summary_large_image", title: "Free Website Uptime Monitoring | pingdan", description },
};

const solution: Solution = {
  path: "/uptime-monitoring",
  eyebrow: "Website uptime monitoring",
  title: "Catch website downtime before your customers do",
  lede: "Check every minute, measure availability and response time, and send actionable downtime alerts through the channels your team already watches.",
  introTitle: "A green status should mean the site really works",
  intro: "Basic uptime checks only prove that a server returned something. pingdan records each response and lets you define what healthy means, so an error page with a 200 status does not quietly pass. When failures cross your threshold, the right people get the status code, error, timestamp, and recovery notice they need to act.",
  benefits: [
    { title: "1-minute availability checks", text: "Choose any interval from one minute to seven days for each website or endpoint." },
    { title: "Response-time history", text: "Track average, p50, p95, minimum, and maximum latency alongside every pass and failure." },
    { title: "Noise-resistant alerts", text: "Require consecutive failures before a monitor is marked down, then notify again when it recovers." },
    { title: "Shareable status pages", text: "Publish the live state and 90-day uptime of selected services for customers and teammates." },
  ],
  workflow: [
    { title: "Add your URL", text: "Paste a public HTTP or HTTPS URL, name the monitor, and select a check interval and timeout." },
    { title: "Define healthy", text: "Choose the expected status and add optional header, body, JSON, or latency assertions." },
    { title: "Route the alert", text: "Attach email, chat, webhook, paging, push, SMS, or incident-management channels." },
  ],
  faq: [
    { question: "What is website uptime monitoring?", answer: "Website uptime monitoring automatically requests a website on a schedule, evaluates whether the response is healthy, records the result, and alerts someone when the site fails or recovers." },
    { question: "How often should an uptime monitor check my site?", answer: "One minute is a practical default for production websites because it keeps average detection time low. Less critical pages and background services can use longer intervals." },
    { question: "Can a website be down while returning 200 OK?", answer: "Yes. Applications sometimes return a branded error page, empty response, or broken JSON with a 200 status. Body and JSON assertions catch these false-positive healthy checks." },
    { question: "Is pingdan uptime monitoring free?", answer: "Yes. pingdan includes unlimited monitors, one-minute checks, assertions, alert channels, history, and status pages without a paid plan or credit card." },
  ],
  guides: [
    { href: "/blog/what-is-uptime-monitoring", title: "What is uptime monitoring?", text: "Understand availability checks, uptime percentages, and why a bare 200 OK is not enough." },
    { href: "/blog/how-often-should-you-monitor", title: "Choosing a monitor interval", text: "Balance detection speed and alert quality for websites, APIs, and background jobs." },
    { href: "/blog/website-down-alerts-email-telegram", title: "Website downtime alerts", text: "Learn what an actionable notification contains and how to choose the right delivery channel." },
  ],
};

export default function UptimeMonitoringPage() { return <SolutionPage solution={solution} />; }
