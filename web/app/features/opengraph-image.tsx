import { renderOgImage, ogSize, ogContentType } from "@/lib/og";

export const runtime = "edge";
export const alt = "Features — pingdan uptime & API monitoring";
export const size = ogSize;
export const contentType = ogContentType;

export default function OpengraphImage() {
  return renderOgImage({
    eyebrow: "Features",
    title: "Assert on everything that matters.",
    sub: "Status, headers, body & JSON path. Response-time charts, uptime history, failure thresholds, and multi-channel alerts.",
  });
}
