"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import { Nav } from "@/components/Nav";
import { channelIcon } from "@/lib/channels";
import { api, displayState, getToken, intervalLabel, daysUntil, sslSeverity, groupStatusColor, monitorTargetSummary, supportsSSLMonitoring, type AlertChannel, type Endpoint, type Group } from "@/lib/api";

const SSL_COLOR: Record<string, string | undefined> = {
  ok: "var(--up)",
  warn: "var(--warn)",
  critical: "var(--down)",
  expired: "var(--down)",
};

/** Compact SSL countdown cell for the endpoints table. */
function SSLCell({ e }: { e: Endpoint }) {
  if (!supportsSSLMonitoring(e)) return <span className="faint">—</span>;
  if (!e.sslExpiresAt) return <span className="faint">{e.sslLastError ? "error" : "—"}</span>;
  const d = daysUntil(e.sslExpiresAt);
  const color = SSL_COLOR[sslSeverity(d)];
  if (d < 0) return <span style={{ color }}>expired</span>;
  return <span style={{ color }}>{d}d</span>;
}

/** Icons for the alert channels attached to an endpoint; label on hover. */
function AlertsCell({ ids, channels }: { ids: string[]; channels: AlertChannel[] }) {
  const attached = ids
    .map((id) => channels.find((c) => c.id === id))
    .filter((c): c is AlertChannel => !!c);
  if (attached.length === 0) return <span className="faint" title="No alert channels attached">—</span>;
  return (
    <span className="row" style={{ gap: 6 }} title={attached.map((c) => c.label).join(", ")}>
      {attached.map((c) => (
        <span key={c.id} aria-label={c.label}>{channelIcon(c.kind)}</span>
      ))}
    </span>
  );
}

export default function EndpointsPage() {
  const router = useRouter();
  const [items, setItems] = useState<Endpoint[]>([]);
  const [groups, setGroups] = useState<Group[]>([]);
  const [channels, setChannels] = useState<AlertChannel[]>([]);
  const [channelMap, setChannelMap] = useState<Record<string, string[]>>({});
  const [loading, setLoading] = useState(true);

  async function refresh() {
    const [eps, grps, chans, chanMap] = await Promise.all([
      api<Endpoint[]>("/endpoints"),
      api<Group[]>("/groups").catch(() => [] as Group[]),
      api<AlertChannel[]>("/alert-channels").catch(() => [] as AlertChannel[]),
      api<Record<string, string[]>>("/endpoints/channels").catch(() => ({}) as Record<string, string[]>),
    ]);
    setItems(eps);
    setGroups(grps);
    setChannels(chans);
    setChannelMap(chanMap);
    setLoading(false);
  }

  const groupName = (id: string | null) => groups.find((g) => g.id === id)?.name ?? null;

  // Bucket endpoints into sections: one per group (in the group list's order),
  // then an "Ungrouped" section last. Empty sections are dropped.
  const sections: { id: string; name: string; items: Endpoint[] }[] = [];
  for (const g of groups) {
    const grouped = items.filter((e) => e.groupId === g.id);
    if (grouped.length > 0) sections.push({ id: g.id, name: g.name, items: grouped });
  }
  const ungrouped = items.filter((e) => !e.groupId || !groupName(e.groupId));
  if (ungrouped.length > 0) sections.push({ id: "__ungrouped__", name: "Ungrouped", items: ungrouped });

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
            <h1>Monitors</h1>
            <div className="subtitle">{items.length} monitor{items.length === 1 ? "" : "s"}</div>
          </div>
          <Link href="/endpoints/new" className="button-link primary">+ New monitor</Link>
        </div>

        {loading ? (
          <p className="muted">Loading…</p>
        ) : items.length === 0 ? (
          <div className="empty">
            <p>No monitors yet.</p>
            <Link href="/endpoints/new" className="button-link primary">Create your first monitor</Link>
          </div>
        ) : (
          <div className="card table-scroll" style={{ padding: 0 }}>
            <table>
              <thead>
                <tr>
                  <th>Name</th>
                  <th>Target</th>
                  <th>State</th>
                  <th>SSL</th>
                  <th>Alerts</th>
                  <th>Interval</th>
                  <th>Last check</th>
                </tr>
              </thead>
              {sections.map((section) => {
                const sectionStates = section.items.map(displayState);
                const down = sectionStates.filter((s) => s === "down").length;
                const degraded = sectionStates.filter((s) => s === "degraded").length;
                const accent = groupStatusColor(sectionStates);
                return (
                  <tbody key={section.id} style={{ ["--group-accent" as string]: accent }}>
                    {sections.length > 1 && (
                      <tr className="group-row">
                        <td colSpan={7}>
                          <span className="group-row-name">{section.name}</span>
                          <span className="group-count">{section.items.length}</span>
                          {down > 0 && <span className="pill down">{down} down</span>}
                          {degraded > 0 && <span className="pill degraded">{degraded} degraded</span>}
                        </td>
                      </tr>
                    )}
                    {section.items.map((e) => {
                      const state = displayState(e);
                      return (
                      <tr
                        key={e.id}
                        style={{ cursor: "pointer" }}
                        onClick={() => router.push(`/endpoints/${e.id}`)}
                      >
                        <td>
                          <div className="row">
                            <span className={`dot ${state}`} />
                            <strong>{e.name}</strong>
                          </div>
                        </td>
                        <td className="mono muted">
                          {monitorTargetSummary(e)}
                        </td>
                        <td>
                          <span
                            className={`pill ${state}`}
                            title={
                              state === "degraded"
                                ? `${e.consecutiveFailures} failed check${e.consecutiveFailures === 1 ? "" : "s"} in a row — alerts at ${e.failureThreshold}`
                                : undefined
                            }
                          >
                            {state}
                          </span>
                        </td>
                        <td className="num mono"><SSLCell e={e} /></td>
                        <td><AlertsCell ids={channelMap[e.id] ?? []} channels={channels} /></td>
                        <td className="num">{intervalLabel(e.intervalSec)}</td>
                        <td className="mono muted">{e.lastCheckedAt ? new Date(e.lastCheckedAt).toLocaleString() : "—"}</td>
                      </tr>
                      );
                    })}
                  </tbody>
                );
              })}
            </table>
          </div>
        )}
      </div>
    </>
  );
}
