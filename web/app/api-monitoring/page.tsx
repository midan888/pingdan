import type { Metadata } from "next";
import { SolutionPage, type Solution } from "@/components/SolutionPage";

const description = "Free API monitoring with 1-minute HTTP checks, status, header, body and JSON path assertions, latency thresholds, history, and instant alerts.";

export const metadata: Metadata = {
  title: "Free API Monitoring with JSON & Response Assertions",
  description,
  keywords: ["API monitoring", "REST API monitoring", "API uptime monitor", "JSON path assertions", "HTTP endpoint monitoring"],
  alternates: { canonical: "/api-monitoring" },
  openGraph: { title: "API Monitoring with Deep Assertions | pingdan", description, url: "/api-monitoring" },
  twitter: { card: "summary_large_image", title: "API Monitoring with Deep Assertions | pingdan", description },
};

const solution: Solution = {
  path: "/api-monitoring",
  eyebrow: "API monitoring",
  title: "Monitor what your API returns, not just whether it responds",
  lede: "Run scheduled HTTP checks and validate status codes, headers, response bodies, JSON paths, and latency on every response.",
  introTitle: "Catch the broken responses that uptime-only checks miss",
  intro: "An API can return 200 OK while its payload is empty, stale, or shaped incorrectly. pingdan evaluates multiple assertions after each request and stores the exact failure detail. You see whether the transport, contract, data, or performance failed instead of starting an incident with a generic red light.",
  benefits: [
    { title: "Deep response assertions", text: "Compare status, response time, headers, raw body content, and dotted JSON paths." },
    { title: "Flexible HTTP requests", text: "Monitor endpoints with GET, POST, PUT, PATCH, DELETE, HEAD, or OPTIONS and a custom timeout." },
    { title: "Latency percentiles", text: "Use average, p50, and p95 response time to spot slow degradation before a hard failure." },
    { title: "Failure-level diagnostics", text: "See the actual status, latency, error, and assertion failure recorded for each check." },
  ],
  workflow: [
    { title: "Choose the endpoint", text: "Add the API URL and HTTP method, then set an interval from one minute to seven days." },
    { title: "Write the contract", text: "Require a status and combine header, body, JSON path, or response-time comparisons." },
    { title: "Act on failures", text: "Use thresholds to filter transient errors and route down and recovery events to your team." },
  ],
  faq: [
    { question: "What should an API monitor validate?", answer: "At minimum, validate the expected HTTP status, a stable field or value in the response, and a response-time ceiling. Critical APIs may also need header and multiple JSON field assertions." },
    { question: "Why is a status-code check not enough?", answer: "Many APIs return 200 for application-level errors or malformed data. A status-only monitor reports healthy even when clients cannot use the response." },
    { question: "Can pingdan monitor JSON responses?", answer: "Yes. JSON path assertions resolve dotted paths, including array indexes such as items.0.id, and compare the resulting value with the target you define." },
    { question: "How quickly can API failures be detected?", answer: "Checks can run once per minute. Your failure threshold controls how many consecutive failures are required before the endpoint is marked down and an alert is sent." },
  ],
  guides: [
    { href: "/blog/api-monitoring-best-practices", title: "9 API monitoring best practices", text: "Build checks that expose contract, dependency, data, and performance failures." },
    { href: "/blog/json-path-assertions-guide", title: "JSON path assertions guide", text: "Validate nested fields and arrays so a plausible but incorrect payload cannot pass." },
    { href: "/blog/http-status-codes-for-monitoring", title: "HTTP status codes for monitoring", text: "Choose the right expected status and decide how redirects and error families should behave." },
  ],
};

export default function ApiMonitoringPage() { return <SolutionPage solution={solution} />; }
