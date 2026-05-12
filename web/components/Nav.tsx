"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { clearToken } from "@/lib/api";

export function Nav() {
  const router = useRouter();
  return (
    <nav className="nav">
      <Link href="/dashboard">Dashboard</Link>
      <Link href="/endpoints">Endpoints</Link>
      <Link href="/channels">Alert channels</Link>
      <div style={{ marginLeft: "auto" }}>
        <button
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
