import Link from "next/link";

const cols = [
  {
    title: "Product",
    links: [
      { href: "/features", label: "Features" },
      { href: "/pricing", label: "Pricing" },
      { href: "/docs", label: "Docs" },
      { href: "/blog", label: "Blog" },
      { href: "/dashboard", label: "Dashboard" },
    ],
  },
  {
    title: "Company",
    links: [
      { href: "/about", label: "About" },
      { href: "/about#contact", label: "Contact" },
    ],
  },
  {
    title: "Get started",
    links: [
      { href: "/register", label: "Create account" },
      { href: "/login", label: "Sign in" },
    ],
  },
];

export function Footer() {
  return (
    <footer className="mkt-footer">
      <div className="mkt-wrap">
        <div className="footer-grid">
          <div>
            <div className="brand">ping<span className="dot">·</span>dan</div>
            <p className="muted" style={{ marginTop: "0.75rem", maxWidth: 260, fontSize: "0.9rem" }}>
              Uptime &amp; API monitoring with deep assertions, response-time charts, and instant alerts.
            </p>
          </div>
          {cols.map((c) => (
            <div className="col" key={c.title}>
              <h4>{c.title}</h4>
              {c.links.map((l) => (
                <Link key={l.href + l.label} href={l.href}>{l.label}</Link>
              ))}
            </div>
          ))}
        </div>
        <div className="footer-bottom">
          <span>© {new Date().getFullYear()} pingdan. All rights reserved.</span>
          <span>Built for engineers who hate downtime.</span>
        </div>
      </div>
    </footer>
  );
}
