import Link from "next/link";
import type { Metadata } from "next";
import type { ReactNode } from "react";
import { MarketingNav } from "@/components/MarketingNav";
import { Footer } from "@/components/Footer";
import { Breadcrumbs } from "@/components/JsonLd";

export const metadata: Metadata = {
  title: "Monitoring Documentation & Setup Guides",
  description:
    "Learn how to set up monitors, write assertions, choose check intervals, and configure alert integrations across email, chat, webhooks, paging, mobile push, SMS and incident tools.",
  alternates: { canonical: "/docs" },
};

const toc = [
  { group: "Getting started", items: [["quickstart", "Quickstart"], ["monitors", "Creating a monitor"], ["intervals", "Check intervals"]] },
  { group: "Validation", items: [["assertions", "Assertions"], ["json-path", "JSON path"]] },
  { group: "Alerting", items: [["thresholds", "Failure thresholds"], ["channels", "Alert channels"], ["channel-guides", "Setup guides"], ["webhooks", "Webhook payload"], ["recipes", "Webhook recipes"]] },
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

type ChannelGuide = {
  id: string;
  name: string;
  bestFor: string;
  env?: ReactNode[];
  fields: ReactNode[];
  steps: ReactNode[];
  example: string;
  docs: { label: string; href: string }[];
};

const channelGuides: ChannelGuide[] = [
  {
    id: "email-guide",
    name: "Email with Resend",
    bestFor: "Durable alerts, audit trails, and teams that already triage incidents from a shared inbox.",
    env: [
      <><code>RESEND_API_KEY</code> — Resend API key with permission to send email.</>,
      <><code>EMAIL_FROM</code> — verified sender such as <code>Pingdan Alerts &lt;alerts@example.com&gt;</code>.</>,
    ],
    fields: [
      <><code>Email address</code> — the recipient, for example <code>oncall@example.com</code>.</>,
    ],
    steps: [
      <>In Resend, add and verify a sending domain.</>,
      <>Create an API key, ideally scoped for sending email.</>,
      <>Set <code>RESEND_API_KEY</code> and <code>EMAIL_FROM</code> on the pingdan API service, then restart it.</>,
      <>Open <Link href="/channels">Alerts</Link>, create an Email channel, enter the recipient, and send a test alert.</>,
    ],
    example: `RESEND_API_KEY=re_xxxxxxxxx
EMAIL_FROM="Pingdan Alerts <alerts@example.com>"

Email address: oncall@example.com`,
    docs: [
      { label: "Resend domains", href: "https://resend.com/docs/dashboard/domains/introduction" },
      { label: "Resend API keys", href: "https://resend.com/docs/api-reference/api-keys/create-api-key" },
    ],
  },
  {
    id: "telegram-guide",
    name: "Telegram",
    bestFor: "Fast mobile push alerts to yourself or a shared on-call group.",
    env: [
      <><code>TELEGRAM_BOT_TOKEN</code> — token from <code>@BotFather</code>.</>,
    ],
    fields: [
      <><code>Telegram chat ID</code> — private chat ID, group ID, or supergroup ID.</>,
    ],
    steps: [
      <>Open Telegram, message <code>@BotFather</code>, run <code>/newbot</code>, and copy the token.</>,
      <>Set <code>TELEGRAM_BOT_TOKEN</code> on the pingdan API service and restart it.</>,
      <>Message the bot once, or add it to the target group and send a message there.</>,
      <>Visit <code>https://api.telegram.org/botYOUR_TOKEN/getUpdates</code> and copy the <code>chat.id</code>. Group IDs are often negative.</>,
      <>Create a Telegram channel in pingdan with that chat ID and send a test alert.</>,
    ],
    example: `TELEGRAM_BOT_TOKEN=123456789:AA...

Telegram chat ID: -1001234567890`,
    docs: [
      { label: "Telegram BotFather tutorial", href: "https://core.telegram.org/bots/tutorial" },
      { label: "Telegram getUpdates", href: "https://core.telegram.org/bots/api#getupdates" },
    ],
  },
  {
    id: "slack-guide",
    name: "Slack",
    bestFor: "Team-visible incident rooms and operations channels.",
    fields: [
      <><code>Slack webhook URL</code> — a URL beginning with <code>https://hooks.slack.com/</code>.</>,
    ],
    steps: [
      <>Create or open a Slack app for your workspace.</>,
      <>Enable Incoming Webhooks for the app.</>,
      <>Add a new webhook to the workspace and choose the channel that should receive alerts.</>,
      <>Copy the webhook URL, create a Slack channel in pingdan, and send a test alert.</>,
    ],
    example: `Slack webhook URL:
https://hooks.slack.com/services/T00000000/B00000000/xxxxxxxx`,
    docs: [
      { label: "Slack incoming webhooks", href: "https://docs.slack.dev/messaging/sending-messages-using-incoming-webhooks/" },
    ],
  },
  {
    id: "discord-guide",
    name: "Discord",
    bestFor: "Community, game, or small-team servers that use Discord channels for operations.",
    fields: [
      <><code>Discord webhook URL</code> — a channel webhook URL from Discord.</>,
    ],
    steps: [
      <>Open the Discord server, go to Server Settings, then Integrations.</>,
      <>Create a webhook, choose the text channel, and give it a recognizable name.</>,
      <>Copy the webhook URL.</>,
      <>Create a Discord channel in pingdan and send a test alert.</>,
    ],
    example: `Discord webhook URL:
https://discord.com/api/webhooks/1234567890/xxxxxxxx`,
    docs: [
      { label: "Discord webhook setup", href: "https://support.discord.com/hc/en-us/articles/228383668-Intro-to-Webhooks" },
      { label: "Discord webhook API", href: "https://docs.discord.com/developers/resources/webhook" },
    ],
  },
  {
    id: "teams-guide",
    name: "Microsoft Teams",
    bestFor: "Microsoft 365 teams that want alerts in a Teams channel or chat.",
    fields: [
      <><code>Power Automate webhook URL</code> — the HTTPS URL generated by a Teams Workflows trigger.</>,
    ],
    steps: [
      <>In Teams, create a Workflow that receives an HTTP request through a webhook URL.</>,
      <>Configure the workflow to post the incoming message or Adaptive Card to the target channel or chat.</>,
      <>Copy the generated webhook URL.</>,
      <>Create a Teams channel in pingdan and send a test alert.</>,
    ],
    example: `Power Automate webhook URL:
https://prod-00.westus.logic.azure.com/workflows/.../triggers/manual/paths/invoke?...`,
    docs: [
      { label: "Teams webhooks with Workflows", href: "https://learn.microsoft.com/en-us/microsoftteams/platform/webhooks-and-connectors/how-to/add-incoming-webhook" },
    ],
  },
  {
    id: "webhook-guide",
    name: "Generic webhook",
    bestFor: "Custom incident automation, internal tools, ticketing glue, and providers with HTTP ingestion.",
    fields: [
      <><code>Webhook URL</code> — any reachable <code>http</code> or <code>https</code> endpoint.</>,
      <><code>Signing secret</code> — optional HMAC secret for verifying pingdan requests.</>,
    ],
    steps: [
      <>Create an endpoint that accepts <code>POST</code> requests with a JSON body.</>,
      <>If you use a signing secret, verify the raw request body against <code>X-Pingdan-Signature</code>.</>,
      <>Create a Generic webhook channel in pingdan with the URL and optional secret.</>,
      <>Send a test alert, then attach the channel to monitors that should trigger automation.</>,
    ],
    example: `Webhook URL: https://example.com/pingdan-alerts
Signing secret: whsec_keep_this_private`,
    docs: [
      { label: "Payload shape below", href: "#webhooks" },
      { label: "Webhook recipes below", href: "#recipes" },
    ],
  },
  {
    id: "pagerduty-guide",
    name: "PagerDuty",
    bestFor: "Production paging, escalation policies, and automatic incident resolution on recovery.",
    fields: [
      <><code>Routing key</code> — the Events API v2 integration key for a PagerDuty service.</>,
    ],
    steps: [
      <>In PagerDuty, create or open the service that should own pingdan incidents.</>,
      <>Add an Events API v2 integration to that service.</>,
      <>Copy the integration key; PagerDuty also calls this the routing key.</>,
      <>Create a PagerDuty channel in pingdan and send a test alert. pingdan triggers and resolves the test incident automatically.</>,
    ],
    example: `Routing key: 0123456789abcdef0123456789abcdef`,
    docs: [
      { label: "PagerDuty Events API v2", href: "https://developer.pagerduty.com/docs/events-api-v2-overview" },
      { label: "Send an alert event", href: "https://developer.pagerduty.com/docs/send-alert-event" },
    ],
  },
  {
    id: "ntfy-guide",
    name: "ntfy",
    bestFor: "Simple mobile push alerts, self-hosted notification topics, or lightweight personal on-call.",
    fields: [
      <><code>Topic</code> — the ntfy topic to publish to.</>,
      <><code>Server</code> — optional server URL; leave blank to use <code>https://ntfy.sh</code>.</>,
      <><code>Access token</code> — optional bearer token for private topics or self-hosted servers.</>,
    ],
    steps: [
      <>Choose a topic name, or create a private topic on your ntfy server.</>,
      <>Subscribe to the topic in the ntfy mobile app, web app, or CLI.</>,
      <>If the topic is protected, create an access token and paste it into pingdan.</>,
      <>Create an ntfy channel in pingdan and send a test alert.</>,
    ],
    example: `Topic: production-alerts
Server: https://ntfy.sh
Access token: tk_xxxxxxxxx`,
    docs: [
      { label: "ntfy publishing", href: "https://docs.ntfy.sh/publish/" },
      { label: "ntfy access tokens", href: "https://docs.ntfy.sh/config/#access-tokens" },
    ],
  },
  {
    id: "pushover-guide",
    name: "Pushover",
    bestFor: "Personal or small-team push alerts across phone, tablet, and desktop clients.",
    env: [
      <><code>PUSHOVER_APP_TOKEN</code> — application token registered in Pushover.</>,
    ],
    fields: [
      <><code>User key</code> — your Pushover user key or a Pushover delivery group key.</>,
    ],
    steps: [
      <>Create or open your Pushover account and install the client on the devices that should receive alerts.</>,
      <>Copy your user key from the Pushover dashboard, or create a delivery group and copy its group key.</>,
      <>Register an application in Pushover and copy its API token.</>,
      <>Set <code>PUSHOVER_APP_TOKEN</code> on the pingdan API service, restart it, then create a Pushover channel and send a test alert.</>,
    ],
    example: `PUSHOVER_APP_TOKEN=po_app_xxxxxxxxx

User key: po_user_xxxxxxxxx`,
    docs: [
      { label: "Pushover API", href: "https://pushover.net/api" },
      { label: "Pushover application token", href: "https://support.pushover.net/i175-how-to-get-a-pushover-api-or-pushover-application-token" },
    ],
  },
  {
    id: "twilio-guide",
    name: "Twilio SMS",
    bestFor: "Last-resort SMS for critical failures where chat or push may be missed.",
    env: [
      <><code>TWILIO_ACCOUNT_SID</code> — Twilio Account SID.</>,
      <><code>TWILIO_AUTH_TOKEN</code> — Twilio Auth Token.</>,
      <><code>TWILIO_FROM</code> — Twilio phone number in E.164 format.</>,
    ],
    fields: [
      <><code>Phone number</code> — recipient in E.164 format, for example <code>+15551234567</code>.</>,
    ],
    steps: [
      <>In Twilio, get an SMS-capable sender number.</>,
      <>Copy the Account SID and Auth Token from the Twilio Console.</>,
      <>Set <code>TWILIO_ACCOUNT_SID</code>, <code>TWILIO_AUTH_TOKEN</code>, and <code>TWILIO_FROM</code> on the pingdan API service, then restart it.</>,
      <>Create a Twilio SMS channel in pingdan with the recipient phone number and send a test alert.</>,
    ],
    example: `TWILIO_ACCOUNT_SID=ACXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX
TWILIO_AUTH_TOKEN=your_auth_token
TWILIO_FROM=+15550000000

Phone number: +15551234567`,
    docs: [
      { label: "Twilio SMS guide", href: "https://www.twilio.com/docs/messaging/tutorials/how-to-send-sms-messages" },
      { label: "Twilio Messages API", href: "https://www.twilio.com/docs/messaging/api/message-resource" },
    ],
  },
  {
    id: "opsgenie-guide",
    name: "Opsgenie",
    bestFor: "Opsgenie alert routing, on-call schedules, and automatic close on recovery.",
    fields: [
      <><code>API key</code> — API integration key with permission to create and close alerts.</>,
      <><code>Region</code> — choose <code>US</code> or <code>EU</code>.</>,
    ],
    steps: [
      <>In Opsgenie, create an API integration for the team or account that should receive alerts.</>,
      <>Copy the integration API key. Prefer an integration key over a general account-management API key.</>,
      <>Choose the matching pingdan region: <code>US</code> for <code>api.opsgenie.com</code>, <code>EU</code> for <code>api.eu.opsgenie.com</code>.</>,
      <>Create an Opsgenie channel in pingdan and send a test alert. pingdan creates and closes the test alert automatically.</>,
    ],
    example: `API key: 01234567-89ab-cdef-0123-456789abcdef
Region: us`,
    docs: [
      { label: "Create Opsgenie API integration", href: "https://support.atlassian.com/opsgenie/docs/create-a-default-api-integration/" },
      { label: "Opsgenie API key management", href: "https://support.atlassian.com/opsgenie/docs/api-key-management/" },
    ],
  },
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

              <h2 id="channel-guides">Channel setup guides</h2>
              <p>
                Each guide ends with the exact value you paste into pingdan. For environment-backed
                channels, set the deployment variables first, restart the API service, then refresh the
                Alerts page so the channel type appears.
              </p>
              <div className="guide-list">
                {channelGuides.map((guide) => (
                  <section className="guide-block" id={guide.id} key={guide.id}>
                    <div className="guide-head">
                      <h3>{guide.name}</h3>
                      <p>{guide.bestFor}</p>
                    </div>
                    <div className="guide-cols">
                      <div>
                        <h4>Deployment env</h4>
                        {guide.env ? (
                          <ul>
                            {guide.env.map((item, i) => <li key={i}>{item}</li>)}
                          </ul>
                        ) : (
                          <p className="muted">None. This channel only needs values from the provider UI.</p>
                        )}
                      </div>
                      <div>
                        <h4>Pingdan fields</h4>
                        <ul>
                          {guide.fields.map((item, i) => <li key={i}>{item}</li>)}
                        </ul>
                      </div>
                    </div>
                    <h4>Setup steps</h4>
                    <ol>
                      {guide.steps.map((item, i) => <li key={i}>{item}</li>)}
                    </ol>
                    <h4>Example</h4>
                    <pre><code>{guide.example}</code></pre>
                    <div className="doc-links">
                      <span>Official docs:</span>
                      {guide.docs.map((doc) => (
                        <a
                          key={doc.href}
                          href={doc.href}
                          target={doc.href.startsWith("http") ? "_blank" : undefined}
                          rel={doc.href.startsWith("http") ? "noreferrer" : undefined}
                        >
                          {doc.label}
                        </a>
                      ))}
                    </div>
                  </section>
                ))}
              </div>

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
