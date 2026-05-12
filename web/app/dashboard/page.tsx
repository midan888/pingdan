"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { Nav } from "@/components/Nav";
import { api, getToken, type Endpoint } from "@/lib/api";

export default function DashboardPage() {
  const router = useRouter();
  const [items, setItems] = useState<Endpoint[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    if (!getToken()) {
      router.replace("/login");
      return;
    }
    api<Endpoint[]>("/endpoints")
      .then(setItems)
      .finally(() => setLoading(false));
  }, [router]);

  const up = items.filter((e) => e.currentState === "up").length;
  const down = items.filter((e) => e.currentState === "down").length;
  const unknown = items.filter((e) => e.currentState === "unknown").length;

  return (
    <>
      <Nav />
      <div className="container">
        <h1>Dashboard</h1>
        {loading ? (
          <p>Loading…</p>
        ) : (
          <div className="grid grid-2">
            <div className="card">
              <h3>Status</h3>
              <p><span className="badge-up">{up} up</span> · <span className="badge-down">{down} down</span> · <span className="badge-unknown">{unknown} unknown</span></p>
            </div>
            <div className="card">
              <h3>Monitored</h3>
              <p>{items.length} endpoint{items.length === 1 ? "" : "s"}</p>
            </div>
          </div>
        )}
      </div>
    </>
  );
}
