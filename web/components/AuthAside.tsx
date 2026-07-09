import Link from "next/link";

const points = [
  "Monitor HTTP endpoints from 1-minute intervals",
  "Assert on status, headers, body & JSON path",
  "Response-time charts and uptime history",
  "Alerts via email, chat, webhooks, paging & SMS",
];

export function AuthAside() {
  return (
    <aside className="auth-aside">
      <Link href="/" className="brand-lg">ping<span className="dot">·</span>dan</Link>

      <div className="pitch">
        <h2>Know the moment your API breaks.</h2>
        <p>Deep uptime monitoring with assertions and charts — set up in under a minute.</p>
        <ul className="checklist">
          {points.map((p) => (
            <li key={p}><span className="tick">✓</span><span>{p}</span></li>
          ))}
        </ul>
      </div>

      <p className="quote">
        “We swapped three tools for pingdan and caught a bad deploy before a single customer noticed.”
      </p>
    </aside>
  );
}
