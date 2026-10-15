import { AUTHOR, SITE_URL, getPosts, smarten } from "@/lib/blog";

// The blog as a feed (D118), for readers and for importing into dev.to.

export const dynamic = "force-static";
export const revalidate = false;

const escape = (text: string) =>
    text
        .replace(/&/g, "&amp;")
        .replace(/</g, "&lt;")
        .replace(/>/g, "&gt;")
        .replace(/"/g, "&quot;");

export function GET() {
    const posts = getPosts();
    const items = posts
        .map((post) => {
            const url = `${SITE_URL}${post.url}`;
            return [
                "<item>",
                `<title>${escape(smarten(post.data.title))}</title>`,
                `<link>${url}</link>`,
                `<guid isPermaLink="true">${url}</guid>`,
                `<pubDate>${post.data.date.toUTCString()}</pubDate>`,
                `<dc:creator>${escape(AUTHOR.name)}</dc:creator>`,
                `<description>${escape(smarten(post.data.description ?? ""))}</description>`,
                "</item>",
            ].join("");
        })
        .join("\n");
    const xml = `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0" xmlns:atom="http://www.w3.org/2005/Atom" xmlns:dc="http://purl.org/dc/elements/1.1/">
<channel>
<title>Cntrl blog</title>
<link>${SITE_URL}/blog</link>
<description>What we ran into running our own machines, and how we fixed it.</description>
<language>en</language>
<atom:link href="${SITE_URL}/blog/rss.xml" rel="self" type="application/rss+xml"/>
${posts[0] ? `<lastBuildDate>${posts[0].data.date.toUTCString()}</lastBuildDate>\n` : ""}${items}
</channel>
</rss>
`;
    return new Response(xml, {
        headers: { "Content-Type": "application/rss+xml; charset=utf-8" },
    });
}
