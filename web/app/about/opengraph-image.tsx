import { renderOgImage, ogSize, ogContentType } from "@/lib/og";

export const runtime = "edge";
export const alt = "About & Contact — pingdan";
export const size = ogSize;
export const contentType = ogContentType;

export default function OpengraphImage() {
  return renderOgImage({
    eyebrow: "About",
    title: "Monitoring built for engineers.",
    sub: "What we believe about uptime & API monitoring — and how to get in touch.",
  });
}
