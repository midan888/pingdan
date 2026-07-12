"use client";

import { Fragment, useEffect, useState } from "react";
import Link from "next/link";
import { usePathname } from "next/navigation";
import { getToken } from "@/lib/api";

const links = [
  { href: "/features", label: "Features" },
  { href: "/uptime-monitoring", label: "Uptime" },
  { href: "/api-monitoring", label: "API" },
  { href: "/pricing", label: "Pricing" },
  { href: "/docs", label: "Docs" },
  { href: "/blog", label: "Blog" },
  { href: "/about", label: "About" },
];

export function MarketingNav() {
  const [authed, setAuthed] = useState(false);
  const [open, setOpen] = useState(false);
  const pathname = usePathname();

  useEffect(() => setAuthed(!!getToken()), []);
  // close the menu whenever the route changes
  useEffect(() => setOpen(false), [pathname]);
  // lock body scroll while the menu is open
  useEffect(() => {
    document.body.style.overflow = open ? "hidden" : "";
    return () => {
      document.body.style.overflow = "";
    };
  }, [open]);

  useEffect(() => {
    if (!open) return;

    const closeOnEscape = (event: KeyboardEvent) => {
      if (event.key === "Escape") setOpen(false);
    };

    window.addEventListener("keydown", closeOnEscape);
    return () => window.removeEventListener("keydown", closeOnEscape);
  }, [open]);

  return (
    <Fragment>
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
              <Link href="/login" className="button-link ghost sign-in">Sign in</Link>
              <Link href="/register" className="button-link primary">Start free</Link>
            </>
          )}
          <button
            className="burger"
            aria-label={open ? "Close menu" : "Open menu"}
            aria-expanded={open}
            aria-controls="mkt-mobile-menu"
            onClick={() => setOpen((o) => !o)}
          >
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" aria-hidden>
              {open ? (
                <path d="M6 6l12 12M18 6L6 18" />
              ) : (
                <path d="M4 7h16M4 12h16M4 17h16" />
              )}
            </svg>
          </button>
        </div>
      </header>
      {open && (
        <div id="mkt-mobile-menu" className="mkt-mobile-menu">
          <nav className="mobile-links">
            {links.map((l) => (
              <Link key={l.href} href={l.href}>{l.label}</Link>
            ))}
          </nav>
          <div className="mobile-cta">
            {authed ? (
              <Link href="/dashboard" className="button-link primary">Go to dashboard</Link>
            ) : (
              <>
                <Link href="/register" className="button-link primary">Start free</Link>
                <Link href="/login" className="button-link">Sign in</Link>
              </>
            )}
          </div>
        </div>
      )}
    </Fragment>
  );
}
