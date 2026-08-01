import Link from "next/link";
import type { Metadata } from "next";
import { MarketingNav } from "@/components/MarketingNav";
import { Footer } from "@/components/Footer";
import { Breadcrumbs } from "@/components/JsonLd";
import { CookieSettingsLink } from "@/components/CookieConsent";

export const metadata: Metadata = {
  title: "Privacy Policy",
  description:
    "What pingdan collects, why we collect it, who we share it with, and how long we keep it. Written in plain English.",
  alternates: { canonical: "/privacy" },
};

/** Shown in the header and in the closing section. */
const LAST_UPDATED = "2 August 2026";

const subprocessors = [
  {
    name: "Resend",
    gets: "Recipient email address and the contents of the message",
    why: "Sends account email (verification, password reset) and email alerts",
  },
  {
    name: "Telegram",
    gets: "The chat ID you configure and the alert text",
    why: "Delivers Telegram alerts",
  },
  {
    name: "Twilio",
    gets: "The phone number you configure and the alert text",
    why: "Delivers SMS alerts",
  },
  {
    name: "Pushover",
    gets: "The user or device key you configure and the alert text",
    why: "Delivers push alerts",
  },
  {
    name: "Google",
    gets: "Your email address, name and avatar URL, returned to us after you approve sign-in",
    why: "Google sign-in",
  },
  {
    name: "GitHub",
    gets: "Your email address, name and avatar URL, returned to us after you approve sign-in",
    why: "GitHub sign-in",
  },
  {
    name: "Google Analytics",
    gets: "Pseudonymous usage data — pages viewed, approximate location, device and browser",
    why: "Understanding how the site is used. Only ever loaded with cookies after you accept",
  },
];

export default function PrivacyPage() {
  return (
    <div className="mkt">
      <Breadcrumbs trail={[{ name: "Privacy Policy", path: "/privacy" }]} />
      <MarketingNav />

      <section className="hero" style={{ padding: "4rem 0 1rem" }}>
        <div className="mkt-wrap">
          <span className="eyebrow">Legal</span>
          <h1 style={{ fontSize: "clamp(2rem,4vw,3rem)" }}>Privacy Policy</h1>
          <p className="lede">
            What we collect, why we collect it, who else sees it, and how long we keep it — in
            plain English, with no clauses designed to be skipped.
          </p>
          <p className="muted" style={{ fontSize: "0.9rem" }}>Last updated {LAST_UPDATED}</p>
        </div>
      </section>

      <section className="mkt-section tight">
        <div className="mkt-wrap" style={{ maxWidth: 760 }}>
          <div className="prose">
            <div className="policy-todo">
              <strong>Before publishing:</strong> replace the bracketed fields below with your
              registered company details. GDPR Article 13 requires the controller to be
              identifiable, so this policy is not complete until they are filled in.
            </div>

            <h2>Who we are</h2>
            <p>
              pingdan is an uptime and API monitoring service operated by{" "}
              <strong>[LEGAL COMPANY NAME]</strong>, a company registered in{" "}
              <strong>[COUNTRY]</strong> under company number{" "}
              <strong>[REGISTRATION NUMBER]</strong>, with its registered office at{" "}
              <strong>[REGISTERED ADDRESS]</strong>. In this policy, &ldquo;we&rdquo; and
              &ldquo;us&rdquo; mean that company. We are the data controller for the personal data
              described here.
            </p>
            <p>
              For anything privacy-related, email{" "}
              <a href="mailto:support@pingdan.dev">support@pingdan.dev</a>.
            </p>

            <h2>The short version</h2>
            <ul>
              <li>We collect what we need to run your monitors and tell you when they fail.</li>
              <li>We do not sell your data, and we do not use it for advertising.</li>
              <li>Analytics cookies are only set if you accept them. Declining changes nothing else.</li>
              <li>pingdan is free, so we never handle payment or card details.</li>
              <li>
                Nothing expires on a timer, and nothing lingers after you delete it — removing a
                monitor takes its full check history with it, immediately.
              </li>
            </ul>

            <h2>What we collect</h2>

            <h3>Account data</h3>
            <p>
              Your email address, and — if you sign in with Google or GitHub — the name and avatar
              URL those providers return, along with the provider&apos;s account identifier. If you
              register with a password instead, we store a bcrypt hash of it. We never store your
              password in a readable form and cannot recover it for you.
            </p>

            <h3>Monitor configuration</h3>
            <p>
              Everything you enter when setting up a check: the name, the URL or host, the HTTP
              method, the check interval and timeout, the failure threshold, and the assertions you
              define. If you point a monitor at a URL that contains an API key or token, that value
              is stored as part of the monitor.
            </p>

            <h3>Check results</h3>
            <p>
              For every check we run we store the status code, the response time in milliseconds,
              whether it passed, any connection error message, and the time it ran.
            </p>
            <p>
              <strong>One thing worth knowing:</strong> when an assertion fails, we store the actual
              value it compared against so you can see why. For a body or JSON-path assertion, that
              value is taken from the response your endpoint returned. If you assert on an endpoint
              that returns personal data, that data can end up in your failure history. Point
              monitors at dedicated health endpoints where you can, and keep assertions narrow.
              Deleting the monitor clears that history immediately.
            </p>

            <h3>Alert channels</h3>
            <p>
              Whatever a channel needs in order to reach you: an email address, a Telegram chat ID,
              a phone number, a Slack, Discord, Teams or custom webhook URL, a PagerDuty or Opsgenie
              key, an ntfy topic, or a Pushover key. These are credentials — treat the ones you paste
              in the same way you would treat any other secret.
            </p>

            <h3>Status pages</h3>
            <p>
              If you publish a status page, its slug, title, description and the monitors you add to
              it are visible to anyone with the link. That is the point of a status page, but it does
              mean you choose what goes on it.
            </p>

            <h3>Technical data</h3>
            <p>
              Our servers keep short-lived logs containing IP addresses and request metadata. We use
              them to keep the service running and to investigate abuse, nothing else.
            </p>

            <h2>Cookies and analytics</h2>
            <p>
              We use Google Analytics 4 to understand how the site is used. It runs in Google&apos;s
              Consent Mode with every storage category denied by default, so no analytics cookie is
              set on your first visit. Nothing is stored on your device for analytics unless you
              press Accept on the cookie banner.
            </p>
            <p>
              You can change your mind at any time — <CookieSettingsLink inline /> reopens the
              banner, and declining is as easy as accepting.
            </p>
            <p>
              Separately, when you sign in we store a session token in your browser&apos;s local
              storage. This is not an analytics cookie and is not optional: without it you would be
              signed out on every page load. Signing out removes it.
            </p>

            <h2>Who we share data with</h2>
            <p>
              We do not sell personal data. We share it only with the providers below, and only the
              parts they need to do their job. Providers are used only where you have configured the
              relevant feature.
            </p>
            <div className="table-scroll">
              <table>
                <thead>
                  <tr>
                    <th>Provider</th>
                    <th>What it receives</th>
                    <th>Why</th>
                  </tr>
                </thead>
                <tbody>
                  {subprocessors.map((p) => (
                    <tr key={p.name}>
                      <td>{p.name}</td>
                      <td>{p.gets}</td>
                      <td>{p.why}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
            <p>
              If you configure a webhook, Slack, Discord, Teams, PagerDuty, Opsgenie or ntfy channel,
              we send alert content to the address you supply. What happens to it after that is
              governed by that service&apos;s privacy policy, not ours.
            </p>
            <p>
              We may also disclose data where the law requires it, or where it is necessary to
              establish or defend a legal claim.
            </p>

            <h2>Where your data is processed</h2>
            <p>
              Our servers and database are located in the European Union. Some of the providers
              listed above process data outside the EU; where they do, those transfers are covered by
              the European Commission&apos;s Standard Contractual Clauses or an equivalent safeguard
              under Chapter V of the GDPR.
            </p>

            <h2>How long we keep it</h2>
            <p>
              We do not expire your data on a timer. Everything you create stays for as long as you
              want it — and the moment you delete it, it is gone.
            </p>
            <ul>
              <li>
                <strong>Monitors, assertions and alert channels</strong> — kept until you delete them.
                Deletion is immediate: the record is removed from the live database as part of the
                request, not queued for later. There is no soft delete and no recycle bin, so we
                cannot undo it for you.
              </li>
              <li>
                <strong>Check results and assertion failure history</strong> — kept for as long as
                the monitor exists. Deleting a monitor deletes its entire history in the same
                operation, including any response data a failed assertion captured along the way.
              </li>
              <li>
                <strong>Status pages</strong> — kept until you delete them, at which point the public
                URL stops resolving.
              </li>
              <li>
                <strong>Server logs</strong> — kept briefly for operational and security purposes,
                then rotated out.
              </li>
            </ul>
            <p>
              To close your account entirely, email{" "}
              <a href="mailto:support@pingdan.dev">support@pingdan.dev</a>. We delete the account
              along with every monitor, check result, alert channel and status page attached to it.
              There is no self-service account deletion in the product yet, so this one goes through
              a human — we will confirm once it is done.
            </p>
            <p className="muted" style={{ fontSize: "0.92rem" }}>
              One caveat, so the word &ldquo;instantly&rdquo; is not overstated: routine database
              backups may hold a copy for a short period after deletion, until those backups are
              rotated out in the normal course. Nothing is restored from them except to recover from
              a failure.
            </p>

            <h2>Our legal basis for using your data</h2>
            <p>Under the GDPR, we rely on:</p>
            <ul>
              <li>
                <strong>Performance of a contract</strong> — running your monitors, sending your
                alerts and maintaining your account. Without this data there is no service.
              </li>
              <li>
                <strong>Legitimate interests</strong> — keeping the service secure, preventing abuse,
                and diagnosing faults.
              </li>
              <li>
                <strong>Consent</strong> — analytics cookies, and only those. You can withdraw it at
                any time without affecting anything else.
              </li>
            </ul>

            <h2>Your rights</h2>
            <p>
              If you are in the EEA or the UK you have the right to access your data, correct it,
              have it deleted, restrict or object to how we use it, receive a portable copy, and
              withdraw consent you have given. Email{" "}
              <a href="mailto:support@pingdan.dev">support@pingdan.dev</a> and we will respond within
              one month.
            </p>
            <p>
              You also have the right to complain to your local data protection authority if you
              think we have handled your data badly. We would rather you told us first so we can put
              it right.
            </p>

            <h2>Security</h2>
            <p>
              Traffic to pingdan is encrypted with TLS. Passwords are hashed with bcrypt. Sessions
              use signed, expiring tokens. Access to the production database is limited to those who
              need it to operate the service. No system is perfect, and we will not pretend
              otherwise — but if we ever discover a breach affecting your data, we will tell you.
            </p>

            <h2>Children</h2>
            <p>
              pingdan is a tool for people running software in production. It is not directed at
              children, and we do not knowingly collect data from anyone under 16. If you believe a
              child has created an account, email us and we will remove it.
            </p>

            <h2>Changes to this policy</h2>
            <p>
              If we change how we handle your data in a way that affects you, we will update this
              page and change the date at the top. This version is dated {LAST_UPDATED}.
            </p>

            <h2>Contact</h2>
            <p>
              Questions about any of this go to{" "}
              <a href="mailto:support@pingdan.dev">support@pingdan.dev</a>. A real person reads it.
            </p>
          </div>
        </div>
      </section>

      <section className="mkt-section">
        <div className="mkt-wrap">
          <div className="cta-band">
            <h2>Monitoring that respects your data</h2>
            <p>Set up your first monitor free. No card, no tracking you didn&apos;t agree to.</p>
            <Link href="/register" className="button-link primary btn-lg">Start free</Link>
          </div>
        </div>
      </section>

      <Footer />
    </div>
  );
}
