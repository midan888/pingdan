"use client";

import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";
import { clearToken } from "@/lib/api";

const links = [
  { href: "/dashboard", label: "Dashboard" },
  { href: "/endpoints", label: "Endpoints" },
  { href: "/groups", label: "Groups" },
  { href: "/channels", label: "Alerts" },
  { href: "/status-pages", label: "Status Pages" },
];

export function Nav() {
  const router = useRouter();
  const pathname = usePathname();
  return (
    <nav className="nav">
      <Link href="/dashboard" className="brand">ping<span className="dot">·</span>dan</Link>
      {links.map((l) => (
        <Link
          key={l.href}
          href={l.href}
          className={pathname?.startsWith(l.href) ? "active" : ""}
        >
          {l.label}
        </Link>
      ))}
      <div style={{ marginLeft: "auto" }}>
        <button
          className="ghost"
          onClick={() => {
            clearToken();
            router.push("/login");
          }}
        >
          Log out
        </button>
      </div>
    </nav>
  );
}
