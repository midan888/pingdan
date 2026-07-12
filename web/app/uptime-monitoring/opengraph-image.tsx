import { renderOgImage, ogSize, ogContentType } from "@/lib/og";

export const runtime = "edge";
export const alt = "pingdan website uptime monitoring";
export const size = ogSize;
export const contentType = ogContentType;

export default function OpengraphImage() {
  return renderOgImage({
    eyebrow: "Website uptime monitoring",
    title: "Catch downtime before your customers do.",
    sub: "1-minute checks, deep response validation, latency history, and instant recovery-aware alerts.",
  });
}
