// Decorative, static product preview used in the marketing hero.
// Deterministic bar heights so server/client render identically (no Math.random).

const services = [
  { name: "api.acme.com", state: "up", uptime: "99.98%", ms: "142 ms", seed: 3 },
  { name: "checkout", state: "up", uptime: "99.95%", ms: "210 ms", seed: 7 },
  { name: "auth-service", state: "down", uptime: "97.10%", ms: "—", seed: 5 },
];

function bars(seed: number, down: boolean) {
  // simple deterministic pseudo-pattern
  return Array.from({ length: 22 }, (_, i) => {
    const h = 30 + ((i * seed * 7) % 60);
    const isDown = down && i > 16;
    return { h, isDown };
  });
}

export function MiniDashboardPreview() {
  return (
    <div className="preview-frame" aria-hidden>
      <div className="bar"><span /><span /><span /></div>
      <div className="body">
        <div className="mini-grid">
          {services.map((s) => (
            <div className="mini-card" key={s.name}>
              <div className="top">
                <span style={{ display: "inline-flex", alignItems: "center", gap: 6, fontWeight: 600 }}>
                  <span className={`dot ${s.state}`} />{s.name}
                </span>
                <span className={`pill ${s.state}`}>{s.state}</span>
              </div>
              <div className="mini-bars">
                {bars(s.seed, s.state === "down").map((b, i) => (
                  <i key={i} className={b.isDown ? "down" : ""} style={{ height: `${b.h}%` }} />
                ))}
              </div>
              <div className="top" style={{ marginTop: "0.6rem", color: "var(--text-dim)" }}>
                <span>{s.uptime} uptime</span>
                <span className="mono">{s.ms}</span>
              </div>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}
