"use client";

import { useCallback, useEffect, useState } from "react";
import Link from "next/link";
import {
  CONSENT_EVENT,
  type ConsentChoice,
  analyticsEnabled,
  getConsent,
  resetConsent,
  setConsent,
} from "@/lib/consent";

/**
 * Cookie banner for the analytics consent choice.
 *
 * Nothing renders until after mount: the answer lives in localStorage, so
 * server HTML cannot know it and rendering the banner during SSR would both
 * mismatch on hydration and flash for visitors who already decided.
 */
export function CookieConsent() {
  const [visible, setVisible] = useState(false);

  useEffect(() => {
    if (!analyticsEnabled) return;

    const sync = () => setVisible(getConsent() === null);
    sync();

    // resetConsent() elsewhere (e.g. a "cookie settings" link) reopens this.
    window.addEventListener(CONSENT_EVENT, sync);
    return () => window.removeEventListener(CONSENT_EVENT, sync);
  }, []);

  const choose = useCallback((choice: ConsentChoice) => {
    setConsent(choice);
    setVisible(false);
  }, []);

  if (!visible) return null;

  return (
    <div className="cookie-banner" role="dialog" aria-label="Cookie consent" aria-live="polite">
      <p>
        We use Google Analytics to understand how pingdan gets used. Analytics cookies are only
        set if you accept — declining keeps everything working, we just stop counting. See our{" "}
        <Link href="/privacy">Privacy Policy</Link>.
      </p>
      <div className="cookie-banner-actions">
        <button type="button" className="button-link ghost" onClick={() => choose("denied")}>
          Decline
        </button>
        <button type="button" className="button-link primary" onClick={() => choose("granted")}>
          Accept
        </button>
      </div>
    </div>
  );
}

/**
 * Entry point for reopening the banner. Withdrawing consent has to be as easy
 * as giving it, so the choice can never be a one-way door.
 *
 * The footer copy is hidden when analytics is switched off, since it would
 * point at nothing. The `inline` variant used in prose always renders — the
 * privacy policy describes the choice either way, and a sentence built around a
 * link that vanished would read as a typo.
 */
export function CookieSettingsLink({ inline = false }: { inline?: boolean }) {
  if (!inline && !analyticsEnabled) return null;

  return (
    <button
      type="button"
      className={inline ? "cookie-settings-inline" : "cookie-settings"}
      onClick={() => resetConsent()}
    >
      Cookie settings
    </button>
  );
}
