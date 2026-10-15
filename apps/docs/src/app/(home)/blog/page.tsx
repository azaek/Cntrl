import {
    AUTHOR,
    SITE_URL,
    formatDate,
    getPosts,
    isoDate,
    readingMinutes,
    smarten,
} from "@/lib/blog";
import { Rss } from "lucide-react";
import type { Metadata } from "next";
import Link from "next/link";
import { notFound } from "next/navigation";

// The blog's index (D118, Editorial): the newest post leads, the rest follow
// by date under "Earlier".

const DESCRIPTION = "Real problems on real machines, and how to fix them.";

export const metadata: Metadata = {
    title: "Blog",
    description: DESCRIPTION,
    alternates: {
        canonical: `${SITE_URL}/blog`,
        types: { "application/rss+xml": `${SITE_URL}/blog/rss.xml` },
    },
    openGraph: {
        type: "website",
        title: "Blog | Cntrl",
        description: DESCRIPTION,
        url: `${SITE_URL}/blog`,
        siteName: "Cntrl",
        images: [
            {
                url: "/og/blog/image.png",
                width: 1200,
                height: 630,
                alt: "The Cntrl blog",
            },
        ],
    },
    twitter: {
        card: "summary_large_image",
        creator: `@${AUTHOR.handle}`,
        title: "Blog | Cntrl",
        description: DESCRIPTION,
        images: ["/og/blog/image.png"],
    },
};

export default async function BlogIndex() {
    const [lead, ...rest] = getPosts();
    // No page until there's something on it.
    if (!lead) notFound();
    return (
        <main className="mx-auto w-full max-w-[720px] flex-1 px-6 pt-16 pb-32 md:pt-24">
            <div className="flex items-end justify-between gap-6">
                <h1 className="font-sg text-5xl font-semibold tracking-[-0.03em]">
                    Blog
                </h1>
                <a
                    href="/blog/rss.xml"
                    className="text-fd-muted-foreground hover:text-fd-foreground -mr-2 inline-flex h-10 items-center gap-1.5 rounded-md px-2 text-sm transition-colors duration-150"
                >
                    <Rss className="size-4" />
                    RSS
                </a>
            </div>
            <p className="text-fd-muted-foreground mt-3 text-pretty">{DESCRIPTION}</p>
            <article className="relative mt-14 md:mt-20">
                <p className="text-fd-muted-foreground text-sm tabular-nums">
                    <time dateTime={isoDate(lead.data.date)}>
                        {formatDate(lead.data.date)}
                    </time>{" "}
                    · {await readingMinutes(lead)} min read
                </p>
                <h2 className="font-sg mt-3 text-3xl leading-[1.1] font-semibold tracking-[-0.02em] text-balance md:text-4xl">
                    <Link
                        href={lead.url}
                        className="hover:text-fd-foreground/75 transition-colors duration-150 after:absolute after:inset-0"
                    >
                        {smarten(lead.data.title)}
                    </Link>
                </h2>
                <p className="text-fd-muted-foreground mt-4 text-lg text-pretty">
                    {smarten(lead.data.description ?? "")}
                </p>
                <p className="mt-5 text-sm font-medium">{AUTHOR.name}</p>
            </article>
            {rest.length > 0 && (
                <section className="mt-16 md:mt-24">
                    <h2 className="text-fd-muted-foreground text-sm font-medium">
                        Earlier
                    </h2>
                    <ol className="mt-4 divide-y border-y">
                        {rest.map((post) => (
                            <li key={post.url} className="relative flex gap-6 py-4">
                                <time
                                    dateTime={isoDate(post.data.date)}
                                    className="text-fd-muted-foreground w-28 shrink-0 pt-0.5 text-sm tabular-nums"
                                >
                                    {formatDate(post.data.date)}
                                </time>
                                <Link
                                    href={post.url}
                                    className="hover:text-fd-foreground/70 font-medium text-balance transition-colors duration-150 after:absolute after:inset-0"
                                >
                                    {smarten(post.data.title)}
                                </Link>
                            </li>
                        ))}
                    </ol>
                </section>
            )}
        </main>
    );
}
