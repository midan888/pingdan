import Link from "next/link";
import type { Metadata } from "next";
import { MarketingNav } from "@/components/MarketingNav";
import { Footer } from "@/components/Footer";
import { Breadcrumbs } from "@/components/JsonLd";

export const metadata: Metadata = {
  title: "Docs — How pingdan works",
  description:
    "Learn how to set up monitors, write assertions on status, headers, body and JSON path, choose check intervals, and configure email & Telegram alerts.",
  alternates: { canonical: "/docs" },
};

const toc = [
  { group: "Getting started", items: [["quickstart", "Quickstart"], ["monitors", "Creating a monitor"], ["intervals", "Check intervals"]] },
  { group: "Validation", items: [["assertions", "Assertions"], ["json-path", "JSON path"]] },
  { group: "Alerting", items: [["thresholds", "Failure thresholds"], ["channels", "Alert channels"]] },
];

export default function DocsPage() {
  return (
    <div className="mkt">
      <Breadcrumbs trail={[{ name: "Docs", path: "/docs" }]} />
      <MarketingNav />

      <section className="mkt-section tight">
        <div className="mkt-wrap">
          <div className="docs-layout">
            <nav className="docs-side">
              {toc.map((g) => (
                <div key={g.group}>
                  <div className="grp">{g.group}</div>
                  {g.items.map(([id, label]) => (
                    <a key={id} href={`#${id}`}>{label}</a>
                  ))}
                </div>
              ))}
            </nav>

            <article className="prose">
              <span className="eyebrow">Documentation</span>
              <h2 id="quickstart">Quickstart</h2>
              <p>
                pingdan monitors HTTP endpoints on a fixed schedule, evaluates a set of assertions against
                each response, and alerts you when something fails. Here&apos;s the whole flow:
              </p>
              <ol>
                <li>Create an account and open <Link href="/dashboard">your dashboard</Link>.</li>
                <li>Add an endpoint: paste a URL and pick the HTTP method.</li>
                <li>Choose a check interval and what makes a check pass.</li>
                <li>Attach an alert channel so you hear about failures.</li>
              </ol>

              <h2 id="monitors">Creating a monitor</h2>
              <p>
                A monitor is one URL we call on a schedule. Each monitor has a request (method + URL),
                a schedule (interval + timeout), and validation rules. Start with the basics:
              </p>
              <pre><code>{`Method:   GET
URL:      https://api.example.com/healthz
Interval: 1 min
Timeout:  10s
Expected: 200`}</code></pre>
              <div className="callout">
                <strong>Tip:</strong> point monitors at a dedicated <code>/healthz</code> endpoint that
                checks your own dependencies (database, cache, downstream APIs).
              </div>

              <h2 id="intervals">Check intervals</h2>
              <p>
                Choose any interval from <code>1 minute</code> up to <code>7 days</code> — pick a number
                and a unit (minutes, hours or days), or use a quick preset like <code>5 min</code>,{" "}
                <code>1 hr</code> or <code>1 day</code>.
              </p>
              <p>The timeout controls how long we wait for a response before recording a failure.</p>

              <h2 id="assertions">Assertions</h2>
              <p>
                Assertions decide whether a check passes. Each one reads a <em>source</em>, applies a
                <em> comparison</em>, and checks against a <em>target</em>. Available sources:
              </p>
              <ul>
                <li><code>status_code</code> — the HTTP response status</li>
                <li><code>response_time</code> — latency in milliseconds</li>
                <li><code>header</code> — a response header by name</li>
                <li><code>body</code> — the full response body</li>
                <li><code>json_path</code> — a value resolved from a JSON body</li>
              </ul>
              <p>Comparisons include equals, not-equals, greater-than, less-than, contains, not-contains and matches (regex).</p>

              <h2 id="json-path">JSON path</h2>
              <p>
                Use dotted paths to drill into a JSON response. Given this body:
              </p>
              <pre><code>{`{ "status": "ok", "items": [ { "id": 42 } ] }`}</code></pre>
              <p>
                The path <code>status</code> resolves to <code>ok</code>, and <code>items.0.id</code>{" "}
                resolves to <code>42</code>. Combine with any comparison, e.g. assert{" "}
                <code>json_path items.0.id equals 42</code>.
              </p>

              <h2 id="thresholds">Failure thresholds</h2>
              <p>
                To avoid alerting on a single transient blip, a monitor only flips to <strong>down</strong>{" "}
                after a configurable number of <em>consecutive</em> failed checks. Set it to <code>1</code>{" "}
                for instant alerts, or higher to filter noise.
              </p>

              <h2 id="channels">Alert channels</h2>
              <p>
                Attach one or more channels to a monitor. When it goes down — and again when it recovers —
                we notify each channel:
              </p>
              <ul>
                <li><strong>Email</strong> — a message to any address you configure.</li>
                <li><strong>Telegram</strong> — a message to a chat via your bot.</li>
              </ul>
              <p>Manage channels from the <Link href="/channels">Alerts</Link> page.</p>
            </article>
          </div>
        </div>
      </section>

      <Footer />
    </div>
  );
}
