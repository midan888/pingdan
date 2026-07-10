"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { Nav } from "@/components/Nav";
import { api, getToken, type AdminStats, type AdminUser } from "@/lib/api";

export default function AdminPage() {
  const router = useRouter();
  const [stats, setStats] = useState<AdminStats | null>(null);
  const [users, setUsers] = useState<AdminUser[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    if (!getToken()) {
      router.replace("/login");
      return;
    }
    Promise.all([api<AdminStats>("/admin/stats"), api<AdminUser[]>("/admin/users")])
      .then(([s, u]) => {
        setStats(s);
        setUsers(u);
        setLoading(false);
      })
      .catch(() => {
        // Non-admins get a 403 — send them back to their dashboard.
        router.replace("/dashboard");
      });
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [router]);

  return (
    <>
      <Nav />
      <div className="container">
        <div className="page-head">
          <div>
            <h1>Admin</h1>
            <div className="subtitle">Everyone using pingdan, at a glance</div>
          </div>
        </div>

        {loading ? (
          <p className="muted">Loading…</p>
        ) : (
          <>
            <div className="grid grid-5" style={{ marginBottom: "1.5rem" }}>
              <div className="card stat">
                <div className="label">Users</div>
                <div className="value">{stats?.userCount ?? "—"}</div>
              </div>
              <div className="card stat">
                <div className="label">Endpoints</div>
                <div className="value">{stats?.endpointCount ?? "—"}</div>
              </div>
            </div>

            <div className="card">
              <div className="table-scroll">
                <table>
                  <thead>
                    <tr>
                      <th style={{ textAlign: "left" }}>Email</th>
                      <th style={{ textAlign: "left" }}>Name</th>
                      <th style={{ textAlign: "left" }}>Provider</th>
                      <th style={{ textAlign: "right" }}>Endpoints</th>
                      <th style={{ textAlign: "right" }}>Signed up</th>
                    </tr>
                  </thead>
                  <tbody>
                    {users.map((u) => (
                      <tr key={u.id}>
                        <td className="mono">{u.email}</td>
                        <td>{u.name ?? <span className="muted">—</span>}</td>
                        <td>{u.provider ?? <span className="muted">—</span>}</td>
                        <td style={{ textAlign: "right" }}>{u.endpointCount}</td>
                        <td className="muted" style={{ textAlign: "right", whiteSpace: "nowrap" }}>
                          {new Date(u.createdAt).toLocaleDateString()}
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </div>
          </>
        )}
      </div>
    </>
  );
}
