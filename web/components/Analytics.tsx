import { Suspense } from "react";
import Script from "next/script";
import { GA_MEASUREMENT_ID, analyticsEnabled, gaBootstrapScript } from "@/lib/consent";
import { PageViewTracker } from "./PageViewTracker";

/**
 * Google Analytics 4. The bootstrap is a plain inline <script> rather than
 * next/script so it parses and runs synchronously ahead of the async gtag.js
 * request — Consent Mode only honours defaults that are already in the
 * dataLayer when the tag loads.
 */
export function Analytics() {
  if (!analyticsEnabled) return null;

  return (
    <>
      {/* eslint-disable-next-line react/no-danger */}
      <script dangerouslySetInnerHTML={{ __html: gaBootstrapScript }} />
      <Script
        src={`https://www.googletagmanager.com/gtag/js?id=${GA_MEASUREMENT_ID}`}
        strategy="afterInteractive"
      />
      {/* useSearchParams needs a boundary or it opts every page out of static rendering. */}
      <Suspense fallback={null}>
        <PageViewTracker />
      </Suspense>
    </>
  );
}
