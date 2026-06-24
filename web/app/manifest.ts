import type { MetadataRoute } from "next";

export default function manifest(): MetadataRoute.Manifest {
  return {
    name: "pingdan — Uptime & API monitoring",
    short_name: "pingdan",
    description:
      "Monitor HTTP endpoints with deep assertions, response-time charts, and instant email & Telegram alerts.",
    start_url: "/",
    display: "standalone",
    background_color: "#0b1120",
    theme_color: "#0b1120",
    icons: [
      { src: "/icon.svg", type: "image/svg+xml", sizes: "any" },
      { src: "/icon.png", type: "image/png", sizes: "512x512" },
    ],
  };
}
