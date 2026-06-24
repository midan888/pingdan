import { ImageResponse } from "next/og";

export const ogSize = { width: 1200, height: 630 };
export const ogContentType = "image/png";

/**
 * Shared OpenGraph card renderer so every route's social preview is visually
 * consistent. `eyebrow` is the small label, `title` the big headline, `sub`
 * the supporting line.
 */
export function renderOgImage({
  eyebrow,
  title,
  sub,
}: {
  eyebrow: string;
  title: string;
  sub: string;
}) {
  return new ImageResponse(
    (
      <div
        style={{
          height: "100%",
          width: "100%",
          display: "flex",
          flexDirection: "column",
          justifyContent: "center",
          padding: "80px",
          background: "linear-gradient(135deg, #0b1120 0%, #111c33 100%)",
          color: "#f8fafc",
          fontFamily: "sans-serif",
        }}
      >
        <div style={{ display: "flex", alignItems: "center", gap: 24 }}>
          <svg width="64" height="64" viewBox="0 0 32 32">
            <rect width="32" height="32" rx="7" fill="#0f1b30" />
            <path
              d="M5 17h5.2l2.4-7.2a1 1 0 0 1 1.9.05l4.1 13 2.3-6.6a1 1 0 0 1 .94-.67H27"
              fill="none"
              stroke="#34d399"
              strokeWidth="2.4"
              strokeLinecap="round"
              strokeLinejoin="round"
            />
          </svg>
          <span style={{ fontSize: 44, fontWeight: 700 }}>pingdan</span>
          <span style={{ fontSize: 26, color: "#34d399", marginLeft: 8, fontWeight: 600 }}>
            {eyebrow}
          </span>
        </div>
        <div style={{ fontSize: 62, fontWeight: 700, marginTop: 44, lineHeight: 1.1, maxWidth: 980 }}>
          {title}
        </div>
        <div style={{ fontSize: 28, color: "#94a3b8", marginTop: 28, maxWidth: 980, lineHeight: 1.4 }}>
          {sub}
        </div>
      </div>
    ),
    { ...ogSize }
  );
}
