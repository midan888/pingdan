"use client";

import { FormEvent, useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { Nav } from "@/components/Nav";
import { api, getToken, type Endpoint } from "@/lib/api";

export default function EndpointsPage() {
  const router = useRouter();
  const [items, setItems] = useState<Endpoint[]>([]);
  const [name, setName] = useState("");
  const [url, setUrl] = useState("");
  const [intervalSec, setIntervalSec] = useState(60);

  async function refresh() {
    setItems(await api<Endpoint[]>("/endpoints"));
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
    await api("/endpoints", {
      method: "POST",
      body: JSON.stringify({ name, url, intervalSec }),
    });
    setName("");
    setUrl("");
    setIntervalSec(60);
    refresh();
  }

  async function remove(id: string) {
    if (!confirm("Delete this endpoint?")) return;
    await api(`/endpoints/${id}`, { method: "DELETE" });
    refresh();
  }

  return (
    <>
      <Nav />
      <div className="container">
        <h1>Endpoints</h1>

        <form className="card" onSubmit={add}>
          <h3>Add endpoint</h3>
          <div className="grid grid-2">
            <input placeholder="Name (e.g. Production API)" value={name} onChange={(e) => setName(e.target.value)} required />
            <input placeholder="https://api.example.com/healthz" value={url} onChange={(e) => setUrl(e.target.value)} required />
            <input type="number" min={10} placeholder="Interval (seconds)" value={intervalSec} onChange={(e) => setIntervalSec(Number(e.target.value))} />
            <button type="submit">Add</button>
          </div>
        </form>

        <div style={{ marginTop: "1.5rem" }}>
          <table>
            <thead>
              <tr>
                <th>Name</th>
                <th>URL</th>
                <th>State</th>
                <th>Interval</th>
                <th>Last check</th>
                <th></th>
              </tr>
            </thead>
            <tbody>
              {items.map((e) => (
                <tr key={e.id}>
                  <td>{e.name}</td>
                  <td><a href={e.url} target="_blank" rel="noreferrer">{e.url}</a></td>
                  <td><span className={`badge-${e.currentState}`}>{e.currentState}</span></td>
                  <td>{e.intervalSec}s</td>
                  <td>{e.lastCheckedAt ? new Date(e.lastCheckedAt).toLocaleString() : "—"}</td>
                  <td><button onClick={() => remove(e.id)}>Delete</button></td>
                </tr>
              ))}
              {items.length === 0 && (
                <tr><td colSpan={6} style={{ color: "#a1a1aa" }}>No endpoints yet.</td></tr>
              )}
            </tbody>
          </table>
        </div>
      </div>
    </>
  );
}
