import fs from "node:fs";
import path from "node:path";
import { marked } from "marked";

/**
 * Blog content lives as Markdown files in `web/content/blog/*.md`, each with a
 * YAML-lite frontmatter block. Add a new article by dropping a `.md` file in
 * that folder — no code changes needed. The filename (sans `.md`) is the slug.
 *
 * Frontmatter we read:
 *   title, description, date (YYYY-MM-DD), author, tags (comma-separated or
 *   [a, b] list), keywords (same), cover (optional), draft (optional bool).
 */

const BLOG_DIR = path.join(process.cwd(), "content", "blog");

export type PostMeta = {
  slug: string;
  title: string;
  description: string;
  date: string; // ISO date (YYYY-MM-DD)
  author: string;
  tags: string[];
  keywords: string[];
  cover?: string;
  readingMinutes: number;
};

export type Post = PostMeta & { html: string };

marked.setOptions({ gfm: true, breaks: false });

/** Split a frontmatter `--- ... ---` block from the markdown body. */
function splitFrontmatter(raw: string): { data: Record<string, string>; body: string } {
  const match = /^---\s*\n([\s\S]*?)\n---\s*\n?/.exec(raw);
  if (!match) return { data: {}, body: raw };
  const data: Record<string, string> = {};
  for (const line of match[1].split("\n")) {
    const idx = line.indexOf(":");
    if (idx === -1) continue;
    const key = line.slice(0, idx).trim();
    let val = line.slice(idx + 1).trim();
    // Strip wrapping quotes.
    if ((val.startsWith('"') && val.endsWith('"')) || (val.startsWith("'") && val.endsWith("'"))) {
      val = val.slice(1, -1);
    }
    data[key] = val;
  }
  return { data, body: raw.slice(match[0].length) };
}

/** Parse a frontmatter value that may be a list: `[a, b]` or `a, b`. */
function toList(val: string | undefined): string[] {
  if (!val) return [];
  const stripped = val.replace(/^\[/, "").replace(/\]$/, "");
  return stripped
    .split(",")
    .map((s) => s.trim().replace(/^["']|["']$/g, ""))
    .filter(Boolean);
}

function estimateReadingMinutes(markdown: string): number {
  const words = markdown.trim().split(/\s+/).length;
  return Math.max(1, Math.round(words / 220));
}

function parseFile(slug: string, raw: string): Post {
  const { data, body } = splitFrontmatter(raw);
  const html = marked.parse(body) as string;
  return {
    slug,
    title: data.title ?? slug,
    description: data.description ?? "",
    date: data.date ?? "1970-01-01",
    author: data.author ?? "pingdan",
    tags: toList(data.tags),
    keywords: toList(data.keywords).length ? toList(data.keywords) : toList(data.tags),
    cover: data.cover || undefined,
    readingMinutes: estimateReadingMinutes(body),
    html,
  };
}

function isDraft(raw: string): boolean {
  const { data } = splitFrontmatter(raw);
  return data.draft === "true" || data.draft === "1";
}

/** All published posts, newest first. */
export function getAllPosts(): PostMeta[] {
  if (!fs.existsSync(BLOG_DIR)) return [];
  return fs
    .readdirSync(BLOG_DIR)
    .filter((f) => f.endsWith(".md"))
    .map((f) => {
      const slug = f.replace(/\.md$/, "");
      const raw = fs.readFileSync(path.join(BLOG_DIR, f), "utf8");
      if (isDraft(raw)) return null;
      const { html, ...meta } = parseFile(slug, raw);
      return meta;
    })
    .filter((p): p is PostMeta => p !== null)
    .sort((a, b) => (a.date < b.date ? 1 : -1));
}

/** A single post by slug, or null if missing/draft. */
export function getPost(slug: string): Post | null {
  const file = path.join(BLOG_DIR, `${slug}.md`);
  if (!fs.existsSync(file)) return null;
  const raw = fs.readFileSync(file, "utf8");
  if (isDraft(raw)) return null;
  return parseFile(slug, raw);
}

export function getAllSlugs(): string[] {
  return getAllPosts().map((p) => p.slug);
}

/** Up to `n` other posts, preferring those sharing a tag. */
export function getRelatedPosts(slug: string, n = 3): PostMeta[] {
  const all = getAllPosts();
  const current = all.find((p) => p.slug === slug);
  const others = all.filter((p) => p.slug !== slug);
  if (!current) return others.slice(0, n);
  const scored = others
    .map((p) => ({ p, score: p.tags.filter((t) => current.tags.includes(t)).length }))
    .sort((a, b) => b.score - a.score || (a.p.date < b.p.date ? 1 : -1));
  return scored.slice(0, n).map((s) => s.p);
}

export function formatDate(iso: string): string {
  const d = new Date(iso + "T00:00:00Z");
  return d.toLocaleDateString("en-US", { year: "numeric", month: "long", day: "numeric", timeZone: "UTC" });
}
