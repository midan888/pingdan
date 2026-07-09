import { renderOgImage, ogSize, ogContentType } from "@/lib/og";

export const runtime = "edge";
export const alt = "Docs — How pingdan works";
export const size = ogSize;
export const contentType = ogContentType;

export default function OpengraphImage() {
  return renderOgImage({
    eyebrow: "Docs",
    title: "How pingdan works.",
    sub: "Set up monitors, write assertions, choose check intervals, and configure multi-channel alerts.",
  });
}
