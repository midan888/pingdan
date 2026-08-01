/**
 * Google Analytics 4 wired to Consent Mode v2.
 *
 * Every non-essential storage category starts out `denied`, so the first page
 * load never writes an analytics cookie. gtag.js still loads and sends
 * cookieless pings; accepting the banner sends a `consent update` that unlocks
 * cookie-based measurement. Declining keeps the denied defaults in place.
 */

const CONSENT_KEY = "pingdan_consent";

/** Public measurement ID — it ships in the page HTML either way. */
export const GA_MEASUREMENT_ID = "G-KSQXP22FXP";

/** Keep local development out of the production property. */
export const analyticsEnabled = process.env.NODE_ENV === "production";

export type ConsentChoice = "granted" | "denied";

/** Fires on the current tab whenever the visitor makes or changes a choice. */
export const CONSENT_EVENT = "pingdan:consent";

declare global {
  interface Window {
    dataLayer?: unknown[];
    gtag?: (...args: unknown[]) => void;
  }
}

export function getConsent(): ConsentChoice | null {
  if (typeof window === "undefined") return null;
  try {
    const stored = window.localStorage.getItem(CONSENT_KEY);
    return stored === "granted" || stored === "denied" ? stored : null;
  } catch {
    // Safari private mode and friends throw on localStorage access.
    return null;
  }
}

/** Persists the choice, tells gtag about it, and notifies the banner. */
export function setConsent(choice: ConsentChoice) {
  try {
    window.localStorage.setItem(CONSENT_KEY, choice);
  } catch {
    // Non-fatal: the choice still applies for this page load.
  }
  window.gtag?.("consent", "update", consentPayload(choice));
  window.dispatchEvent(new CustomEvent(CONSENT_EVENT, { detail: choice }));
}

/** Clears the stored choice so the banner asks again. */
export function resetConsent() {
  try {
    window.localStorage.removeItem(CONSENT_KEY);
  } catch {
    // Nothing to clear.
  }
  window.dispatchEvent(new CustomEvent(CONSENT_EVENT, { detail: null }));
}

/**
 * The Consent Mode v2 signal set. `security_storage` stays granted — it covers
 * fraud prevention and is not something the visitor opts into.
 */
function consentPayload(choice: ConsentChoice) {
  return {
    ad_storage: choice,
    ad_user_data: choice,
    ad_personalization: choice,
    analytics_storage: choice,
    functionality_storage: choice,
    personalization_storage: choice,
    security_storage: "granted",
  };
}

/**
 * Inline bootstrap, rendered as a blocking <script> ahead of gtag.js.
 *
 * It has to run before gtag.js so the denied defaults are in the dataLayer
 * first; a returning visitor's stored grant is replayed here too, so their
 * first pageview of the session is not measured under the denied default.
 * `gtag('config')` runs here as well, guaranteeing it is queued before the
 * page_view events that PageViewTracker sends after hydration.
 */
export const gaBootstrapScript = `
window.dataLayer = window.dataLayer || [];
function gtag(){dataLayer.push(arguments);}
window.gtag = gtag;
gtag('consent', 'default', ${JSON.stringify({
  ...consentPayload("denied"),
  wait_for_update: 500,
})});
try {
  if (window.localStorage.getItem('${CONSENT_KEY}') === 'granted') {
    gtag('consent', 'update', ${JSON.stringify(consentPayload("granted"))});
  }
} catch (e) {}
gtag('js', new Date());
gtag('config', '${GA_MEASUREMENT_ID}', { send_page_view: false });
`.trim();
