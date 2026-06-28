"use client";

import { FormEvent, useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import { Nav } from "@/components/Nav";
import { api, getToken, type StatusPage } from "@/lib/api";

export default function StatusPagesPage() {
  const router = useRouter();
  const [pages, setPages] = useState<StatusPage[]>([]);
  const [loading, setLoading] = useState(true);

  const [title, setTitle] = useState("");
  const [slug, setSlug] = useState("");
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);

  async function refresh() {
    const p = await api<StatusPage[]>("/status-pages");
    setPages(p);
    setLoading(false);
  }

  useEffect(() => {
    if (!getToken()) {
      router.replace("/login");
      return;
    }
    refresh().catch(() => setLoading(false));
  }, [router]);

  async function add(ev: FormEvent) {
    ev.preventDefault();
    const t = title.trim();
    if (!t) return;
    setSaving(true);
    setError(null);
    try {
      const created = await api<StatusPage>("/status-pages", {
        method: "POST",
        body: JSON.stringify({ title: t, slug: slug.trim() }),
      });
      setTitle("");
      setSlug("");
      router.push(`/status-pages/${created.id}`);
    } catch (e) {
      setError(e instanceof Error ? e.message : "Failed to create status page");
      setSaving(false);
    }
  }

  return (
    <>
      <Nav />
      <div className="container">
        <div className="page-head">
          <div>
            <h1>Status Pages</h1>
            <div className="subtitle">Public, shareable pages showing the live status of a selection of your endpoints.</div>
          </div>
        </div>

        <div className="chan-grid">
          {/* Create form */}
          <form className="card" onSubmit={add}>
            <h3>New status page</h3>
            <div className="field">
              <label>Title</label>
              <input
                placeholder="Acme Status"
                value={title}
                onChange={(e) => { setTitle(e.target.value); if (error) setError(null); }}
                required
                autoFocus
              />
              <div className="hint">Shown publicly at the top of the page.</div>
            </div>
            <div className="field">
              <label>Slug <span className="faint">(optional)</span></label>
              <input
                placeholder="acme"
                value={slug}
                onChange={(e) => { setSlug(e.target.value); if (error) setError(null); }}
              />
              <div className="hint">Public URL path. Leave blank to derive from the title.</div>
            </div>
            {error && <p className="error-text">{error}</p>}
            <button type="submit" className="primary" disabled={saving} style={{ width: "100%" }}>
              {saving ? "Creating…" : "Create status page"}
            </button>
          </form>

          {/* List */}
          <div>
            {loading ? (
              <p className="muted">Loading…</p>
            ) : pages.length === 0 ? (
              <div className="empty">
                <p>No status pages yet.</p>
                <p className="faint" style={{ margin: 0 }}>Create one on the left, then choose which endpoints appear on it.</p>
              </div>
            ) : (
              <div className="stack" style={{ gap: "0.6rem" }}>
                {pages.map((p) => (
                  <div className="card" key={p.id} style={{ padding: "0.9rem 1.1rem" }}>
                    <div className="spread">
                      <div>
                        <strong>{p.title}</strong>
                        <div className="mono muted" style={{ fontSize: "0.82rem" }}>/status/{p.slug}</div>
                      </div>
                      <div className="row" style={{ gap: "0.4rem" }}>
                        <Link href={`/status/${p.slug}`} target="_blank"><button>View</button></Link>
                        <Link href={`/status-pages/${p.id}`}><button className="primary">Edit</button></Link>
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>
        </div>
      </div>
    </>
  );
}
