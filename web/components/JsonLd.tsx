const siteUrl = process.env.NEXT_PUBLIC_SITE_URL ?? "https://pingdan.dev";

/**
 * Renders a JSON-LD structured-data block. The object is serialized server-side,
 * so search engines see it in the initial HTML.
 */
export function JsonLd({ data }: { data: Record<string, unknown> }) {
  return (
    <script
      type="application/ld+json"
      // eslint-disable-next-line react/no-danger
      dangerouslySetInnerHTML={{ __html: JSON.stringify(data) }}
    />
  );
}

/**
 * BreadcrumbList JSON-LD. Pass the trail from the home page down to the current
 * page, e.g. [{ name: "Pricing", path: "/pricing" }]. "Home" is prepended.
 */
export function Breadcrumbs({ trail }: { trail: { name: string; path: string }[] }) {
  const items = [{ name: "Home", path: "/" }, ...trail];
  return (
    <JsonLd
      data={{
        "@context": "https://schema.org",
        "@type": "BreadcrumbList",
        itemListElement: items.map((item, i) => ({
          "@type": "ListItem",
          position: i + 1,
          name: item.name,
          item: `${siteUrl}${item.path}`,
        })),
      }}
    />
  );
}
