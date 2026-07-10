"use client";

import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";
import { useEffect, useState } from "react";
import { api, clearToken, getToken, type Me } from "@/lib/api";

const baseLinks = [
  { href: "/dashboard", label: "Dashboard", icon: "dashboard" },
  { href: "/endpoints", label: "Endpoints", icon: "endpoints" },
  { href: "/groups", label: "Groups", icon: "groups" },
  { href: "/channels", label: "Alerts", icon: "alerts" },
  { href: "/status-pages", label: "Status", icon: "status" },
] as const;

const adminLink = { href: "/admin", label: "Admin", icon: "admin" } as const;

type NavLink = (typeof baseLinks)[number] | typeof adminLink;

function TabIcon({ name }: { name: NavLink["icon"] }) {
  const common = {
    width: 22,
    height: 22,
    viewBox: "0 0 24 24",
    fill: "none",
    stroke: "currentColor",
    strokeWidth: 1.8,
    strokeLinecap: "round" as const,
    strokeLinejoin: "round" as const,
    "aria-hidden": true,
  };
  switch (name) {
    case "dashboard":
      return (
        <svg {...common}>
          <rect x="3" y="3" width="7.5" height="7.5" rx="1.5" />
          <rect x="13.5" y="3" width="7.5" height="7.5" rx="1.5" />
          <rect x="3" y="13.5" width="7.5" height="7.5" rx="1.5" />
          <rect x="13.5" y="13.5" width="7.5" height="7.5" rx="1.5" />
        </svg>
      );
    case "endpoints":
      return (
        <svg {...common}>
          <path d="M3 12h4l2.5-6 4 12 2.5-6h5" />
        </svg>
      );
    case "groups":
      return (
        <svg {...common}>
          <path d="M3 7a2 2 0 0 1 2-2h4l2 2.5h8a2 2 0 0 1 2 2V17a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V7z" />
        </svg>
      );
    case "alerts":
      return (
        <svg {...common}>
          <path d="M18 9a6 6 0 1 0-12 0c0 5-2 6-2 6h16s-2-1-2-6" />
          <path d="M10.3 19a2 2 0 0 0 3.4 0" />
        </svg>
      );
    case "status":
      return (
        <svg {...common}>
          <rect x="3" y="4" width="18" height="13" rx="2" />
          <path d="M7 10.5h2l1.5-3 2.5 5 1.5-2h2.5" />
          <path d="M9 21h6" />
        </svg>
      );
    case "admin":
      return (
        <svg {...common}>
          <path d="M12 3l7 3v5c0 4.5-3 8.5-7 10-4-1.5-7-5.5-7-10V6l7-3z" />
        </svg>
      );
  }
}

export function Nav() {
  const router = useRouter();
  const pathname = usePathname();
  const [isAdmin, setIsAdmin] = useState(false);
  useEffect(() => {
    if (!getToken()) return;
    api<Me>("/me")
      .then((me) => setIsAdmin(me.isAdmin))
      .catch(() => {});
  }, []);
  const links: readonly NavLink[] = isAdmin ? [...baseLinks, adminLink] : baseLinks;
  const logout = () => {
    clearToken();
    router.push("/login");
  };
  return (
    <>
      <nav className="nav">
        <Link href="/dashboard" className="brand">ping<span className="dot">·</span>dan</Link>
        <div className="nav-links">
          {links.map((l) => (
            <Link
              key={l.href}
              href={l.href}
              className={pathname?.startsWith(l.href) ? "active" : ""}
            >
              {l.label}
            </Link>
          ))}
        </div>
        <div style={{ marginLeft: "auto" }}>
          <button className="ghost" onClick={logout}>
            Log out
          </button>
        </div>
      </nav>
      <nav className="tabbar" aria-label="Primary">
        {links.map((l) => (
          <Link
            key={l.href}
            href={l.href}
            className={pathname?.startsWith(l.href) ? "tab active" : "tab"}
            aria-current={pathname?.startsWith(l.href) ? "page" : undefined}
          >
            <TabIcon name={l.icon} />
            <span>{l.label}</span>
          </Link>
        ))}
      </nav>
    </>
  );
}
