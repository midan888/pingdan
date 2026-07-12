import { renderOgImage, ogSize, ogContentType } from "@/lib/og";

export const runtime = "edge";
export const alt = "pingdan SSL certificate expiry monitoring";
export const size = ogSize;
export const contentType = ogContentType;

export default function OpengraphImage() {
  return renderOgImage({
    eyebrow: "SSL certificate monitoring",
    title: "Renew every certificate before it becomes an outage.",
    sub: "Automatic HTTPS certificate inspection, expiry visibility, and alerts while there is still time to act.",
  });
}
