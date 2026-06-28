"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import { Nav } from "@/components/Nav";
import { api, getToken, intervalLabel, daysUntil, sslSeverity, groupColor, type Endpoint, type Group } from "@/lib/api";

const SSL_COLOR: Record<string, string | undefined> = {
  ok: "var(--up)",
  warn: "var(--warn)",
  critical: "var(--down)",
  expired: "var(--down)",
};

/** Compact SSL countdown cell for the endpoints table. */
function SSLCell({ e }: { e: Endpoint }) {
  if (!e.url.startsWith("https://")) return <span className="faint">—</span>;
  if (!e.sslExpiresAt) return <span className="faint">{e.sslLastError ? "error" : "—"}</span>;
  const d = daysUntil(e.sslExpiresAt);
  const color = SSL_COLOR[sslSeverity(d)];
  if (d < 0) return <span style={{ color }}>expired</span>;
  return <span style={{ color }}>{d}d</span>;
}

export default function EndpointsPage() {
  const router = useRouter();
  const [items, setItems] = useState<Endpoint[]>([]);
  const [groups, setGroups] = useState<Group[]>([]);
  const [loading, setLoading] = useState(true);

  async function refresh() {
    const [eps, grps] = await Promise.all([
      api<Endpoint[]>("/endpoints"),
      api<Group[]>("/groups").catch(() => [] as Group[]),
    ]);
    setItems(eps);
    setGroups(grps);
    setLoading(false);
  }

  const groupName = (id: string | null) => groups.find((g) => g.id === id)?.name ?? null;

  // Bucket endpoints into sections: one per group (in the group list's order),
  // then an "Ungrouped" section last. Empty sections are dropped.
  const sections: { id: string; name: string; items: Endpoint[] }[] = [];
  for (const g of groups) {
    const grouped = items.filter((e) => e.groupId === g.id);
    if (grouped.length > 0) sections.push({ id: g.id, name: g.name, items: grouped });
  }
  const ungrouped = items.filter((e) => !e.groupId || !groupName(e.groupId));
  if (ungrouped.length > 0) sections.push({ id: "__ungrouped__", name: "Ungrouped", items: ungrouped });

  useEffect(() => {
    if (!getToken()) {
      router.replace("/login");
      return;
    }
    refresh().catch(() => setLoading(false));
  }, [router]);

  return (
    <>
      <Nav />
      <div className="container">
        <div className="page-head">
          <div>
            <h1>Endpoints</h1>
            <div className="subtitle">{items.length} monitored endpoint{items.length === 1 ? "" : "s"}</div>
          </div>
          <Link href="/endpoints/new"><button className="primary">+ New endpoint</button></Link>
        </div>

        {loading ? (
          <p className="muted">Loading…</p>
        ) : items.length === 0 ? (
          <div className="empty">
            <p>No endpoints yet.</p>
            <Link href="/endpoints/new"><button className="primary">Create your first endpoint</button></Link>
          </div>
        ) : (
          <div className="card" style={{ padding: 0 }}>
            <table>
              <thead>
                <tr>
                  <th>Name</th>
                  <th>URL</th>
                  <th>State</th>
                  <th>SSL</th>
                  <th>Interval</th>
                  <th>Last check</th>
                </tr>
              </thead>
              {sections.map((section, i) => {
                const down = section.items.filter((e) => e.currentState === "down").length;
                const accent = section.id === "__ungrouped__" ? "var(--unknown)" : groupColor(i);
                return (
                  <tbody key={section.id} style={{ ["--group-accent" as string]: accent }}>
                    {sections.length > 1 && (
                      <tr className="group-row">
                        <td colSpan={6}>
                          <span className="group-row-name">{section.name}</span>
                          <span className="group-count">{section.items.length}</span>
                          {down > 0 && <span className="pill down">{down} down</span>}
                        </td>
                      </tr>
                    )}
                    {section.items.map((e) => (
                      <tr
                        key={e.id}
                        style={{ cursor: "pointer" }}
                        onClick={() => router.push(`/endpoints/${e.id}`)}
                      >
                        <td>
                          <div className="row">
                            <span className={`dot ${e.currentState}`} />
                            <strong>{e.name}</strong>
                          </div>
                        </td>
                        <td className="mono muted">{e.method} {e.url}</td>
                        <td><span className={`pill ${e.currentState}`}>{e.currentState}</span></td>
                        <td className="num mono"><SSLCell e={e} /></td>
                        <td className="num">{intervalLabel(e.intervalSec)}</td>
                        <td className="mono muted">{e.lastCheckedAt ? new Date(e.lastCheckedAt).toLocaleString() : "—"}</td>
                      </tr>
                    ))}
                  </tbody>
                );
              })}
            </table>
          </div>
        )}
      </div>
    </>
  );
}
