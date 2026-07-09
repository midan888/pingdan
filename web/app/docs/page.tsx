import Link from "next/link";
import type { Metadata } from "next";
import { MarketingNav } from "@/components/MarketingNav";
import { Footer } from "@/components/Footer";
import { Breadcrumbs } from "@/components/JsonLd";

export const metadata: Metadata = {
  title: "Docs — How pingdan works",
  description:
    "Learn how to set up monitors, write assertions, choose check intervals, and configure alert integrations across email, chat, webhooks, paging, mobile push, SMS and incident tools.",
  alternates: { canonical: "/docs" },
};

const toc = [
  { group: "Getting started", items: [["quickstart", "Quickstart"], ["monitors", "Creating a monitor"], ["intervals", "Check intervals"]] },
  { group: "Validation", items: [["assertions", "Assertions"], ["json-path", "JSON path"]] },
  { group: "Alerting", items: [["thresholds", "Failure thresholds"], ["channels", "Alert channels"], ["webhooks", "Webhook payload"], ["recipes", "Webhook recipes"]] },
];

const channels = [
  ["Email", "Add an address. The deployment must have RESEND_API_KEY set, and EMAIL_FROM should be a verified sender."],
  ["Telegram", "Paste a chat ID. The deployment must have TELEGRAM_BOT_TOKEN set; use getUpdates after messaging the bot to find your chat ID."],
  ["Slack", "Create an incoming webhook and paste the hooks.slack.com URL."],
  ["Discord", "Create a channel webhook and paste the discord.com/api/webhooks URL."],
  ["Microsoft Teams", "Create a Power Automate workflow with a Teams webhook trigger and paste the HTTPS workflow URL."],
  ["Generic webhook", "Paste an HTTP(S) URL. Optionally add a signing secret for HMAC verification."],
  ["PagerDuty", "Create an Events API v2 integration and paste its routing key."],
  ["ntfy", "Enter a topic, optionally a custom server URL and bearer token. The default server is https://ntfy.sh."],
  ["Pushover", "Paste a user or group key. The deployment must have PUSHOVER_APP_TOKEN set."],
  ["Twilio SMS", "Paste an E.164 phone number such as +15551234567. The deployment must have TWILIO_ACCOUNT_SID, TWILIO_AUTH_TOKEN and TWILIO_FROM set."],
  ["Opsgenie", "Paste an API integration key and choose the US or EU region."],
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
                pingdan notifies each attached channel. SSL-expiry warnings use the same channel set.
              </p>
              <ul>
                {channels.map(([name, desc]) => (
                  <li key={name}><strong>{name}</strong> — {desc}</li>
                ))}
              </ul>
              <p>
                Manage channels from the <Link href="/channels">Alerts</Link> page. Use <em>Send test</em>{" "}
                after adding credentials so delivery errors surface before an incident.{" "}
                Channels that need deployment credentials only appear in the create form after those
                environment variables are configured.
              </p>

              <h2 id="webhooks">Webhook payload</h2>
              <p>
                Generic webhooks receive the full structured alert as JSON. A typical endpoint-down
                payload looks like this:
              </p>
              <pre><code>{`{
  "event": "endpoint.down",
  "endpoint": {
    "id": "9d4b2d3c-...",
    "name": "Production API",
    "url": "https://api.example.com/healthz"
  },
  "check": {
    "statusCode": 503,
    "checkedAt": "2026-07-10T08:30:00Z"
  },
  "subject": "[pingdan] Production API — DOWN",
  "body": "Endpoint: Production API\\nURL: https://api.example.com/healthz\\nState: DOWN\\nStatus: 503\\nAt: 2026-07-10T08:30:00Z"
}`}</code></pre>
              <p>
                SSL alerts use <code>event: "ssl.expiring"</code> and include an <code>ssl</code> object
                with <code>daysLeft</code> and <code>expiresAt</code>. Recovery alerts use{" "}
                <code>event: "endpoint.recovered"</code>.
              </p>
              <p>
                If you set a webhook secret, pingdan signs the exact request body with HMAC-SHA256 and
                sends <code>X-Pingdan-Signature: sha256=&lt;hex&gt;</code>. In Node.js, verification is:
              </p>
              <pre><code>{`import crypto from "node:crypto";

const expected = "sha256=" + crypto
  .createHmac("sha256", process.env.PINGDAN_WEBHOOK_SECRET)
  .update(rawRequestBody)
  .digest("hex");

const received = Buffer.from(signature ?? "");
const trusted = Buffer.from(expected);
if (received.length !== trusted.length || !crypto.timingSafeEqual(received, trusted)) {
  throw new Error("bad signature");
}`}</code></pre>

              <h2 id="recipes">Webhook recipes</h2>
              <p>
                Tools with flexible webhook ingestion can use the generic webhook channel directly:
              </p>
              <ul>
                <li>
                  <strong>Better Stack</strong> — create an incoming webhook integration in Better Stack,
                  copy its ingest URL, add a pingdan Generic webhook channel, and attach it to monitors.
                  Use the payload&apos;s <code>event</code> and <code>endpoint.id</code> fields for routing
                  or deduplication rules.
                </li>
                <li>
                  <strong>Grafana OnCall</strong> — create an HTTP webhook integration, paste the generated
                  URL into a pingdan Generic webhook channel, and map <code>endpoint.down</code> to firing
                  and <code>endpoint.recovered</code> to resolved in the integration template.
                </li>
              </ul>
            </article>
          </div>
        </div>
      </section>

      <Footer />
    </div>
  );
}
