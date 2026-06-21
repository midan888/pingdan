"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import { Nav } from "@/components/Nav";
import { api, getToken, intervalLabel, type Endpoint } from "@/lib/api";

export default function EndpointsPage() {
  const router = useRouter();
  const [items, setItems] = useState<Endpoint[]>([]);
  const [loading, setLoading] = useState(true);

  async function refresh() {
    setItems(await api<Endpoint[]>("/endpoints"));
    setLoading(false);
  }

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
                  <th>Interval</th>
                  <th>Last check</th>
                </tr>
              </thead>
              <tbody>
                {items.map((e) => (
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
                    <td className="num">{intervalLabel(e.intervalSec)}</td>
                    <td className="mono muted">{e.lastCheckedAt ? new Date(e.lastCheckedAt).toLocaleString() : "—"}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>
    </>
  );
}
