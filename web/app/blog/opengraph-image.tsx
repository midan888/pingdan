import { renderOgImage, ogSize, ogContentType } from "@/lib/og";

export const runtime = "edge";
export const alt = "pingdan Blog — Uptime & API monitoring guides";
export const size = ogSize;
export const contentType = ogContentType;

export default function OpengraphImage() {
  return renderOgImage({
    eyebrow: "Blog",
    title: "Monitoring, alerting & reliability.",
    sub: "Field-tested guides on uptime monitoring, API health checks, alerting and SLAs.",
  });
}
