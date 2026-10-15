import { loader } from "fumadocs-core/source";
import { toFumadocsSource } from "fumadocs-mdx/runtime/server";
import { blog as posts } from "fumadocs-mdx:collections/server";

// The blog (D118): MDX in content/blog, at /blog.
export const blog = loader({
    baseUrl: "/blog",
    source: toFumadocsSource(posts, []),
});

export type Post = ReturnType<typeof blog.getPages>[number];

export const SITE_URL = "https://www.cntrl.pw";

/** Who writes the articles, linked to the X profile that posts them. */
export const AUTHOR = {
    name: "AzaeK",
    handle: "datbugdied",
    url: "https://x.com/datbugdied",
    avatar: "https://github.com/azaek.png",
    bio: "Builds cntrl, and writes up what breaks along the way.",
} as const;

const published = (post: Post) =>
    process.env.NODE_ENV !== "production" || !post.data.draft;

/** The posts, newest first; drafts only in development. */
export function getPosts(): Post[] {
    return blog
        .getPages()
        .filter(published)
        .sort((a, b) => b.data.date.getTime() - a.data.date.getTime());
}

/** One post by its slug, unless it's a draft outside development. */
export function getPost(slug: string): Post | undefined {
    const post = blog.getPage([slug]);
    return post && published(post) ? post : undefined;
}

/** Minutes to read a post, at 230 words a minute. */
export async function readingMinutes(post: Post): Promise<number> {
    const text = await post.data.getText("processed");
    return Math.max(1, Math.round(text.split(/\s+/).filter(Boolean).length / 230));
}

const MONTHS = [
    "Jan",
    "Feb",
    "Mar",
    "Apr",
    "May",
    "Jun",
    "Jul",
    "Aug",
    "Sep",
    "Oct",
    "Nov",
    "Dec",
];

/** "16 Oct 2026": day first, three letters for every month. */
export function formatDate(date: Date): string {
    return `${date.getUTCDate()} ${MONTHS[date.getUTCMonth()]} ${date.getUTCFullYear()}`;
}

export function isoDate(date: Date): string {
    return date.toISOString().slice(0, 10);
}

/** Curly quotes for frontmatter text, which doesn't go through the MDX prose. */
export function smarten(text: string): string {
    return text
        .replace(/(\w)'(\w)/g, "$1’$2")
        .replace(/(^|[\s(\[])"/g, "$1“")
        .replace(/"/g, "”")
        .replace(/(^|[\s(\[])'/g, "$1‘")
        .replace(/'/g, "’");
}
