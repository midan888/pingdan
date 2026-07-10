"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { getToken } from "@/lib/api";

const links = [
  { href: "/features", label: "Features" },
  { href: "/pricing", label: "Pricing" },
  { href: "/docs", label: "Docs" },
  { href: "/blog", label: "Blog" },
  { href: "/about", label: "About" },
];

export function MarketingNav() {
  const [authed, setAuthed] = useState(false);
  useEffect(() => setAuthed(!!getToken()), []);

  return (
    <header className="mkt-nav">
      <Link href="/" className="brand">ping<span className="dot">·</span>dan</Link>
      <nav className="links">
        {links.map((l) => (
          <Link key={l.href} href={l.href}>{l.label}</Link>
        ))}
      </nav>
      <div className="right">
        {authed ? (
          <Link href="/dashboard" className="button-link primary">Go to dashboard</Link>
        ) : (
          <>
            <Link href="/login" className="button-link ghost">Sign in</Link>
            <Link href="/register" className="button-link primary">Start free</Link>
          </>
        )}
      </div>
    </header>
  );
}
