"use client";

import { FormEvent, useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { Nav } from "@/components/Nav";
import { api, getToken, type AlertChannel } from "@/lib/api";

type Kind = "email" | "telegram";

const KINDS: { value: Kind; label: string; icon: string; hint: string; placeholder: string }[] = [
  {
    value: "email",
    label: "Email",
    icon: "✉",
    hint: "We'll send alerts to this address when an endpoint goes down or recovers.",
    placeholder: "oncall@company.com",
  },
  {
    value: "telegram",
    label: "Telegram",
    icon: "✈",
    hint: "Message @userinfobot on Telegram to get your chat ID, then add our bot to that chat.",
    placeholder: "123456789",
  },
];

function validate(kind: Kind, value: string): string | null {
  const v = value.trim();
  if (!v) return "This field is required.";
  if (kind === "email" && !/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(v)) return "Enter a valid email address.";
  if (kind === "telegram" && !/^-?\d+$/.test(v)) return "Chat ID should be a number (e.g. 123456789).";
  return null;
}

export default function ChannelsPage() {
  const router = useRouter();
  const [items, setItems] = useState<AlertChannel[]>([]);
  const [loading, setLoading] = useState(true);

  const [kind, setKind] = useState<Kind>("email");
  const [label, setLabel] = useState("");
  const [value, setValue] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [saving, setSaving] = useState(false);

  const active = KINDS.find((k) => k.value === kind)!;

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
    const err = validate(kind, value);
    if (err) {
      setError(err);
      return;
    }
    setError(null);
    setSaving(true);
    try {
      const config = kind === "email" ? { to: value.trim() } : { chatId: value.trim() };
      await api("/alert-channels", {
        method: "POST",
        body: JSON.stringify({ kind, label: label.trim(), config }),
      });
      setLabel("");
      setValue("");
      await refresh();
    } catch (e) {
      setError(e instanceof Error ? e.message : "Failed to add channel");
    } finally {
      setSaving(false);
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
                {KINDS.map((k) => (
                  <button
                    type="button"
                    key={k.value}
                    className={`chip ${kind === k.value ? "selected" : ""}`}
                    onClick={() => { setKind(k.value); setError(null); }}
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

            <div className="field">
              <label>{kind === "email" ? "Email address" : "Telegram chat ID"}</label>
              <input
                placeholder={active.placeholder}
                value={value}
                onChange={(e) => { setValue(e.target.value); if (error) setError(null); }}
                inputMode={kind === "telegram" ? "numeric" : "email"}
                required
              />
              <div className="hint">{active.hint}</div>
            </div>

            {error && <p className="error-text">{error}</p>}

            <button type="submit" className="primary" style={{ width: "100%", marginTop: "0.25rem" }} disabled={saving}>
              {saving ? "Adding…" : "Add channel"}
            </button>
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
                  const meta = KINDS.find((k) => k.value === c.kind);
                  const target = c.kind === "email" ? (c.config.to as string) : (c.config.chatId as string);
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
                        <button className="danger" onClick={() => remove(c.id, c.label)}>Delete</button>
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
