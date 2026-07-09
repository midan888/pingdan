"use client";

import { FormEvent, useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { Nav } from "@/components/Nav";
import { api, getToken, type AlertChannel, type AlertChannelKind } from "@/lib/api";
import {
  CHANNEL_KINDS,
  channelKind,
  channelTarget,
  configForKind,
  validateChannelConfig,
} from "@/lib/channels";

export default function ChannelsPage() {
  const router = useRouter();
  const [items, setItems] = useState<AlertChannel[]>([]);
  const [loading, setLoading] = useState(true);

  const [kind, setKind] = useState<AlertChannelKind>("email");
  const [label, setLabel] = useState("");
  const [values, setValues] = useState<Record<string, string>>({});
  const [error, setError] = useState<string | null>(null);
  const [saving, setSaving] = useState(false);
  const [testing, setTesting] = useState(false);
  const [formNote, setFormNote] = useState<string | null>(null);
  const [testingId, setTestingId] = useState<string | null>(null);

  const active = channelKind(kind);

  function updateValue(key: string, value: string) {
    setValues((prev) => ({ ...prev, [key]: value }));
    if (error) setError(null);
    if (formNote) setFormNote(null);
  }

  async function refresh() {
    setItems(await api<AlertChannel[]>("/alert-channels"));
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
    const err = validateChannelConfig(kind, values);
    if (err) {
      setError(err);
      return;
    }
    setError(null);
    setSaving(true);
    try {
      await api("/alert-channels", {
        method: "POST",
        body: JSON.stringify({ kind, label: label.trim(), config: configForKind(kind, values) }),
      });
      setLabel("");
      setValues({});
      await refresh();
    } catch (e) {
      setError(e instanceof Error ? e.message : "Failed to add channel");
    } finally {
      setSaving(false);
    }
  }

  async function sendTestForm() {
    const err = validateChannelConfig(kind, values);
    if (err) {
      setError(err);
      return;
    }
    setError(null);
    setFormNote(null);
    setTesting(true);
    try {
      await api("/alert-channels/test", {
        method: "POST",
        body: JSON.stringify({ kind, config: configForKind(kind, values) }),
      });
      setFormNote("Test alert sent — check the destination.");
    } catch (e) {
      setError(e instanceof Error ? e.message : "Failed to send test alert");
    } finally {
      setTesting(false);
    }
  }

  async function testChannel(c: AlertChannel) {
    setTestingId(c.id);
    try {
      await api("/alert-channels/test", {
        method: "POST",
        body: JSON.stringify({ kind: c.kind, config: c.config }),
      });
      alert(`Test alert sent to "${c.label}". Check the destination.`);
    } catch (e) {
      alert(`Test failed: ${e instanceof Error ? e.message : "unknown error"}`);
    } finally {
      setTestingId(null);
    }
  }

  async function remove(id: string, channelLabel: string) {
    if (!confirm(`Delete the channel "${channelLabel}"? Endpoints using it will stop alerting here.`)) return;
    await api(`/alert-channels/${id}`, { method: "DELETE" });
    refresh();
  }

  return (
    <>
      <Nav />
      <div className="container">
        <div className="page-head">
          <div>
            <h1>Alert channels</h1>
            <div className="subtitle">Where pingdan sends notifications when an endpoint changes state.</div>
          </div>
        </div>

        <div className="chan-grid">
          {/* Add form */}
          <form className="card" onSubmit={add}>
            <h3>Add a channel</h3>

            <div className="field">
              <label>Type</label>
              <div className="chips">
                {CHANNEL_KINDS.map((k) => (
                  <button
                    type="button"
                    key={k.value}
                    className={`chip ${kind === k.value ? "selected" : ""}`}
                    onClick={() => { setKind(k.value); setError(null); setFormNote(null); }}
                  >
                    <span aria-hidden>{k.icon}</span> {k.label}
                  </button>
                ))}
              </div>
            </div>

            <div className="field">
              <label>Label</label>
              <input
                placeholder="On-call team"
                value={label}
                onChange={(e) => setLabel(e.target.value)}
                required
              />
              <div className="hint">A name to recognise this channel by.</div>
            </div>

            {active.fields.map((field) => (
              <div className="field" key={field.key}>
                <label>
                  {field.label}
                  {field.optional && <span className="faint"> (optional)</span>}
                </label>
                <input
                  type={field.sensitive ? "password" : "text"}
                  placeholder={field.placeholder}
                  value={values[field.key] ?? ""}
                  onChange={(e) => updateValue(field.key, e.target.value)}
                  inputMode={field.inputMode}
                  required={!field.optional}
                />
                {field.hint && <div className="hint">{field.hint}</div>}
              </div>
            ))}

            {error && <p className="error-text">{error}</p>}
            {formNote && <p className="hint" style={{ color: "var(--ok, #16a34a)" }}>{formNote}</p>}

            <div className="row" style={{ gap: "0.5rem", marginTop: "0.25rem" }}>
              <button
                type="button"
                onClick={sendTestForm}
                disabled={testing || saving}
                style={{ flex: "0 0 auto" }}
              >
                {testing ? "Sending…" : "Send test"}
              </button>
              <button type="submit" className="primary" style={{ flex: 1 }} disabled={saving || testing}>
                {saving ? "Adding…" : "Add channel"}
              </button>
            </div>
          </form>

          {/* List */}
          <div>
            {loading ? (
              <p className="muted">Loading…</p>
            ) : items.length === 0 ? (
              <div className="empty">
                <p>No alert channels yet.</p>
                <p className="faint" style={{ margin: 0 }}>Add one on the left, then attach it to your endpoints.</p>
              </div>
            ) : (
              <div className="stack" style={{ gap: "0.6rem" }}>
                {items.map((c) => {
                  const meta = channelKind(c.kind);
                  const target = channelTarget(c);
                  return (
                    <div className="card" key={c.id} style={{ padding: "0.9rem 1.1rem" }}>
                      <div className="spread">
                        <div className="row">
                          <div className="chan-ico" aria-hidden>{meta?.icon ?? "•"}</div>
                          <div>
                            <strong>{c.label}</strong>
                            <div className="mono muted" style={{ fontSize: "0.82rem" }}>
                              <span className="pill unknown" style={{ marginRight: "0.5rem" }}>{c.kind}</span>
                              {target}
                            </div>
                          </div>
                        </div>
                        <div className="row" style={{ gap: "0.4rem" }}>
                          <button onClick={() => testChannel(c)} disabled={testingId === c.id}>
                            {testingId === c.id ? "Testing…" : "Test"}
                          </button>
                          <button className="danger" onClick={() => remove(c.id, c.label)}>Delete</button>
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
