import { renderOgImage, ogSize, ogContentType } from "@/lib/og";
import { getPost, getAllSlugs } from "@/lib/blog";

// Node runtime (not edge): we read post frontmatter from the filesystem.
export const size = ogSize;
export const contentType = ogContentType;
export const alt = "pingdan Blog article";

export function generateStaticParams() {
  return getAllSlugs().map((slug) => ({ slug }));
}

export default function OpengraphImage({ params }: { params: { slug: string } }) {
  const post = getPost(params.slug);
  return renderOgImage({
    eyebrow: "Blog",
    title: post?.title ?? "pingdan Blog",
    sub: post?.description ?? "Uptime & API monitoring guides.",
  });
}
