import { renderOgImage, ogSize, ogContentType } from "@/lib/og";

export const runtime = "edge";
export const alt = "pingdan — Uptime & API monitoring with deep assertions";
export const size = ogSize;
export const contentType = ogContentType;

export default function OpengraphImage() {
  return renderOgImage({
    eyebrow: "Uptime & API monitoring",
    title: "Know the moment your API breaks.",
    sub: "Deep assertions on status, headers, body & JSON path. Response-time charts and instant email & Telegram alerts. Free — every feature.",
  });
}
