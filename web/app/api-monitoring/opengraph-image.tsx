import { renderOgImage, ogSize, ogContentType } from "@/lib/og";

export const runtime = "edge";
export const alt = "pingdan API monitoring with deep response assertions";
export const size = ogSize;
export const contentType = ogContentType;

export default function OpengraphImage() {
  return renderOgImage({
    eyebrow: "API monitoring",
    title: "Monitor what your API actually returns.",
    sub: "Assert on status, headers, body, JSON paths, and latency on every scheduled HTTP check.",
  });
}
