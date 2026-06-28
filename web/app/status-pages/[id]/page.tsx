"use client";

import { useEffect, useState } from "react";
import { useParams, useRouter } from "next/navigation";
import Link from "next/link";
import { Nav } from "@/components/Nav";
import {
  api,
  getToken,
  type Endpoint,
  type StatusPage,
  type StatusPageItem,
} from "@/lib/api";

type Detail = { page: StatusPage; items: StatusPageItem[] };

// Editor row state: whether the endpoint is included and its public name override.
type RowState = { included: boolean; displayName: string };

export default function StatusPageEditor() {
  const router = useRouter();
  const params = useParams();
  const id = params.id as string;

  const [page, setPage] = useState<StatusPage | null>(null);
  const [endpoints, setEndpoints] = useState<Endpoint[]>([]);
  const [rows, setRows] = useState<Record<string, RowState>>({});
  const [title, setTitle] = useState("");
  const [slug, setSlug] = useState("");
  const [description, setDescription] = useState("");

  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [notFound, setNotFound] = useState(false);
  const [savedAt, setSavedAt] = useState<number | null>(null);

  useEffect(() => {
    if (!getToken()) {
      router.replace("/login");
      return;
    }
    (async () => {
      try {
        const [detail, eps] = await Promise.all([
          api<Detail>(`/status-pages/${id}`),
          api<Endpoint[]>("/endpoints").catch(() => [] as Endpoint[]),
        ]);
        setPage(detail.page);
        setTitle(detail.page.title);
        setSlug(detail.page.slug);
        setDescription(detail.page.description);
        setEndpoints(eps);
        const initial: Record<string, RowState> = {};
        for (const e of eps) initial[e.id] = { included: false, displayName: "" };
        for (const it of detail.items) {
          initial[it.endpointId] = { included: true, displayName: it.displayName ?? "" };
        }
        setRows(initial);
      } catch (e) {
        if (e instanceof Error && e.message === "not found") setNotFound(true);
        else setError(e instanceof Error ? e.message : "Failed to load");
      } finally {
        setLoading(false);
      }
    })();
  }, [id, router]);

  function toggle(epId: string) {
    setRows((r) => ({ ...r, [epId]: { ...r[epId], included: !r[epId].included } }));
  }
  function setName(epId: string, name: string) {
    setRows((r) => ({ ...r, [epId]: { ...r[epId], displayName: name } }));
  }

  async function save() {
    const t = title.trim();
    if (!t) { setError("Title is required"); return; }
    setSaving(true);
    setError(null);
    try {
      await api(`/status-pages/${id}`, {
        method: "PUT",
        body: JSON.stringify({ title: t, slug: slug.trim(), description: description.trim() }),
      });
      // Items in endpoint list order; only the included ones are sent.
      const items = endpoints
        .filter((e) => rows[e.id]?.included)
        .map((e) => ({ endpointId: e.id, displayName: rows[e.id].displayName.trim() || null }));
      await api(`/status-pages/${id}/items`, { method: "PUT", body: JSON.stringify(items) });
      const fresh = await api<Detail>(`/status-pages/${id}`);
      setPage(fresh.page);
      setSlug(fresh.page.slug);
      setSavedAt(Date.now());
    } catch (e) {
      setError(e instanceof Error ? e.message : "Failed to save");
    } finally {
      setSaving(false);
    }
  }

  async function remove() {
    if (!page) return;
    if (!confirm(`Delete the status page "${page.title}"? Its public URL will stop working.`)) return;
    await api(`/status-pages/${id}`, { method: "DELETE" });
    router.push("/status-pages");
  }

  if (notFound) {
    return (
      <>
        <Nav />
        <div className="container">
          <div className="empty">
            <p>Status page not found.</p>
            <Link href="/status-pages"><button>Back to status pages</button></Link>
          </div>
        </div>
      </>
    );
  }

  const publicPath = `/status/${page?.slug ?? slug}`;
  const includedCount = endpoints.filter((e) => rows[e.id]?.included).length;

  return (
    <>
      <Nav />
      <div className="container">
        <div className="page-head">
          <div>
            <h1>{loading ? "Status page" : page?.title}</h1>
            <div className="subtitle">Configure what visitors see at your public status page.</div>
          </div>
          <div className="row" style={{ gap: "0.6rem" }}>
            <Link href="/status-pages"><button>Back</button></Link>
            {page && <Link href={publicPath} target="_blank"><button>View public page</button></Link>}
          </div>
        </div>

        {loading ? (
          <p className="muted">Loading…</p>
        ) : (
          <div className="stack" style={{ gap: "1.25rem" }}>
            {page && (
              <div className="card" style={{ padding: "0.9rem 1.1rem" }}>
                <div className="spread">
                  <div>
                    <div className="label">Public URL</div>
                    <div className="mono" style={{ fontSize: "0.9rem" }}>{publicPath}</div>
                  </div>
                  <button
                    onClick={() => {
                      const full = `${window.location.origin}${publicPath}`;
                      navigator.clipboard?.writeText(full);
                    }}
                  >
                    Copy link
                  </button>
                </div>
              </div>
            )}

            {/* Page details */}
            <div className="card">
              <h3>Details</h3>
              <div className="field">
                <label>Title</label>
                <input value={title} onChange={(e) => setTitle(e.target.value)} required />
              </div>
              <div className="field">
                <label>Slug</label>
                <input value={slug} onChange={(e) => setSlug(e.target.value)} />
                <div className="hint">Lowercase letters, numbers and dashes. Public URL path.</div>
              </div>
              <div className="field">
                <label>Description <span className="faint">(optional)</span></label>
                <textarea
                  value={description}
                  onChange={(e) => setDescription(e.target.value)}
                  rows={2}
                  placeholder="Live status of our services."
                />
              </div>
            </div>

            {/* Endpoint selection */}
            <div className="card">
              <div className="spread" style={{ marginBottom: "0.75rem" }}>
                <h3 style={{ margin: 0 }}>Endpoints</h3>
                <span className="muted" style={{ fontSize: "0.82rem" }}>{includedCount} included</span>
              </div>
              {endpoints.length === 0 ? (
                <p className="faint" style={{ margin: 0 }}>
                  No endpoints yet. <Link href="/endpoints/new">Create one</Link> first.
                </p>
              ) : (
                <div className="stack" style={{ gap: "0.5rem" }}>
                  {endpoints.map((e) => {
                    const row = rows[e.id] ?? { included: false, displayName: "" };
                    return (
                      <div
                        key={e.id}
                        className="row"
                        style={{ gap: "0.75rem", padding: "0.5rem 0", borderBottom: "1px solid var(--border)" }}
                      >
                        <input
                          type="checkbox"
                          checked={row.included}
                          onChange={() => toggle(e.id)}
                          aria-label={`Include ${e.name}`}
                          style={{ width: "auto" }}
                        />
                        <span className={`dot ${e.currentState}`} />
                        <div style={{ minWidth: 0, flex: "0 0 200px" }}>
                          <strong style={{ whiteSpace: "nowrap", overflow: "hidden", textOverflow: "ellipsis", display: "block" }}>{e.name}</strong>
                          <div className="mono faint" style={{ fontSize: "0.72rem", whiteSpace: "nowrap", overflow: "hidden", textOverflow: "ellipsis" }}>{e.url}</div>
                        </div>
                        <input
                          placeholder={`Public name (defaults to "${e.name}")`}
                          value={row.displayName}
                          onChange={(ev) => setName(e.id, ev.target.value)}
                          disabled={!row.included}
                          style={{ flex: 1 }}
                        />
                      </div>
                    );
                  })}
                </div>
              )}
            </div>

            {error && <p className="error-text">{error}</p>}

            <div className="spread">
              <button className="danger" onClick={remove}>Delete page</button>
              <div className="row" style={{ gap: "0.6rem" }}>
                {savedAt && <span className="muted" style={{ fontSize: "0.82rem" }}>Saved ✓</span>}
                <button className="primary" onClick={save} disabled={saving}>
                  {saving ? "Saving…" : "Save changes"}
                </button>
              </div>
            </div>
          </div>
        )}
      </div>
    </>
  );
}
