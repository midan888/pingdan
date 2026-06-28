"use client";

import { FormEvent, useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { Nav } from "@/components/Nav";
import { api, getToken, type Endpoint, type Group } from "@/lib/api";

export default function GroupsPage() {
  const router = useRouter();
  const [groups, setGroups] = useState<Group[]>([]);
  const [endpoints, setEndpoints] = useState<Endpoint[]>([]);
  const [loading, setLoading] = useState(true);

  const [name, setName] = useState("");
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const [editingId, setEditingId] = useState<string | null>(null);
  const [editName, setEditName] = useState("");

  async function refresh() {
    const [g, e] = await Promise.all([
      api<Group[]>("/groups"),
      api<Endpoint[]>("/endpoints").catch(() => [] as Endpoint[]),
    ]);
    setGroups(g);
    setEndpoints(e);
    setLoading(false);
  }

  useEffect(() => {
    if (!getToken()) {
      router.replace("/login");
      return;
    }
    refresh().catch(() => setLoading(false));
  }, [router]);

  function countFor(groupId: string) {
    return endpoints.filter((e) => e.groupId === groupId).length;
  }

  async function add(ev: FormEvent) {
    ev.preventDefault();
    const trimmed = name.trim();
    if (!trimmed) return;
    setSaving(true);
    setError(null);
    try {
      await api("/groups", { method: "POST", body: JSON.stringify({ name: trimmed }) });
      setName("");
      await refresh();
    } catch (e) {
      setError(e instanceof Error ? e.message : "Failed to create group");
    } finally {
      setSaving(false);
    }
  }

  function startEdit(g: Group) {
    setEditingId(g.id);
    setEditName(g.name);
  }

  async function saveEdit(id: string) {
    const trimmed = editName.trim();
    if (!trimmed) return;
    try {
      await api(`/groups/${id}`, { method: "PUT", body: JSON.stringify({ name: trimmed }) });
      setEditingId(null);
      await refresh();
    } catch (e) {
      alert(e instanceof Error ? e.message : "Failed to rename group");
    }
  }

  async function remove(g: Group) {
    const n = countFor(g.id);
    const note = n > 0 ? ` Its ${n} endpoint${n === 1 ? "" : "s"} will become ungrouped (not deleted).` : "";
    if (!confirm(`Delete the group "${g.name}"?${note}`)) return;
    await api(`/groups/${g.id}`, { method: "DELETE" });
    refresh();
  }

  return (
    <>
      <Nav />
      <div className="container">
        <div className="page-head">
          <div>
            <h1>Groups</h1>
            <div className="subtitle">Organize endpoints into named collections. Optional — endpoints can stay ungrouped.</div>
          </div>
        </div>

        <div className="chan-grid">
          {/* Add form */}
          <form className="card" onSubmit={add}>
            <h3>New group</h3>
            <div className="field">
              <label>Group name</label>
              <input
                placeholder="Production"
                value={name}
                onChange={(e) => { setName(e.target.value); if (error) setError(null); }}
                required
                autoFocus
              />
              <div className="hint">e.g. &ldquo;Production&rdquo;, &ldquo;Internal APIs&rdquo;, or a client name.</div>
            </div>
            {error && <p className="error-text">{error}</p>}
            <button type="submit" className="primary" disabled={saving} style={{ width: "100%" }}>
              {saving ? "Creating…" : "Create group"}
            </button>
          </form>

          {/* List */}
          <div>
            {loading ? (
              <p className="muted">Loading…</p>
            ) : groups.length === 0 ? (
              <div className="empty">
                <p>No groups yet.</p>
                <p className="faint" style={{ margin: 0 }}>Create one on the left, then assign endpoints to it from the endpoint form.</p>
              </div>
            ) : (
              <div className="stack" style={{ gap: "0.6rem" }}>
                {groups.map((g) => {
                  const n = countFor(g.id);
                  const editing = editingId === g.id;
                  return (
                    <div className="card" key={g.id} style={{ padding: "0.9rem 1.1rem" }}>
                      <div className="spread">
                        {editing ? (
                          <input
                            value={editName}
                            onChange={(e) => setEditName(e.target.value)}
                            onKeyDown={(e) => {
                              if (e.key === "Enter") saveEdit(g.id);
                              if (e.key === "Escape") setEditingId(null);
                            }}
                            autoFocus
                            style={{ flex: 1, marginRight: "0.6rem" }}
                          />
                        ) : (
                          <div>
                            <strong>{g.name}</strong>
                            <div className="muted" style={{ fontSize: "0.82rem" }}>
                              {n} endpoint{n === 1 ? "" : "s"}
                            </div>
                          </div>
                        )}
                        <div className="row" style={{ gap: "0.4rem" }}>
                          {editing ? (
                            <>
                              <button className="primary" onClick={() => saveEdit(g.id)}>Save</button>
                              <button onClick={() => setEditingId(null)}>Cancel</button>
                            </>
                          ) : (
                            <>
                              <button onClick={() => startEdit(g)}>Rename</button>
                              <button className="danger" onClick={() => remove(g)}>Delete</button>
                            </>
                          )}
                        </div>
                      </div>
                    </div>
                  );
                })}
              </div>
            )}
          </div>
        </div>
      </div>
    </>
  );
}
