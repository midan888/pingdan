import { renderOgImage, ogSize, ogContentType } from "@/lib/og";

export const runtime = "edge";
export const alt = "Pricing — pingdan is free, every feature included";
export const size = ogSize;
export const contentType = ogContentType;

export default function OpengraphImage() {
  return renderOgImage({
    eyebrow: "Pricing",
    title: "It's free. All of it.",
    sub: "Unlimited monitors, 1-minute intervals, all alert channels, and full history. No limits, no credit card.",
  });
}
