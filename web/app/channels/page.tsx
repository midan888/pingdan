"use client";

import { FormEvent, useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { Nav } from "@/components/Nav";
import { api, getToken, type AlertChannel } from "@/lib/api";

export default function ChannelsPage() {
  const router = useRouter();
  const [items, setItems] = useState<AlertChannel[]>([]);
  const [kind, setKind] = useState<"email" | "telegram">("email");
  const [label, setLabel] = useState("");
  const [value, setValue] = useState("");

  async function refresh() {
    setItems(await api<AlertChannel[]>("/alert-channels"));
  }

  useEffect(() => {
    if (!getToken()) {
      router.replace("/login");
      return;
    }
    refresh();
  }, [router]);

  async function add(ev: FormEvent) {
    ev.preventDefault();
    const config = kind === "email" ? { to: value } : { chatId: value };
    await api("/alert-channels", {
      method: "POST",
      body: JSON.stringify({ kind, label, config }),
    });
    setLabel("");
    setValue("");
    refresh();
  }

  async function remove(id: string) {
    await api(`/alert-channels/${id}`, { method: "DELETE" });
    refresh();
  }

  return (
    <>
      <Nav />
      <div className="container">
        <h1>Alert channels</h1>

        <form className="card" onSubmit={add}>
          <h3>Add channel</h3>
          <div className="grid grid-2">
            <select value={kind} onChange={(e) => setKind(e.target.value as "email" | "telegram")}>
              <option value="email">Email</option>
              <option value="telegram">Telegram</option>
            </select>
            <input placeholder="Label (e.g. Oncall email)" value={label} onChange={(e) => setLabel(e.target.value)} required />
            <input placeholder={kind === "email" ? "you@example.com" : "Telegram chat ID"} value={value} onChange={(e) => setValue(e.target.value)} required />
            <button type="submit">Add</button>
          </div>
        </form>

        <div style={{ marginTop: "1.5rem" }}>
          <table>
            <thead>
              <tr>
                <th>Label</th>
                <th>Kind</th>
                <th>Target</th>
                <th></th>
              </tr>
            </thead>
            <tbody>
              {items.map((c) => (
                <tr key={c.id}>
                  <td>{c.label}</td>
                  <td>{c.kind}</td>
                  <td>{c.kind === "email" ? (c.config.to as string) : (c.config.chatId as string)}</td>
                  <td><button onClick={() => remove(c.id)}>Delete</button></td>
                </tr>
              ))}
              {items.length === 0 && (
                <tr><td colSpan={4} style={{ color: "#a1a1aa" }}>No channels yet.</td></tr>
              )}
            </tbody>
          </table>
        </div>
      </div>
    </>
  );
}
