import { AuthorLink } from "@/components/blog/author";
import {
    AUTHOR,
    SITE_URL,
    formatDate,
    getPost,
    getPosts,
    isoDate,
    readingMinutes,
    smarten,
} from "@/lib/blog";
import { getMDXComponents } from "@/mdx-components";
import type { Metadata } from "next";
import Link from "next/link";
import { notFound } from "next/navigation";

// An article (D118, Editorial): one centered column about 68 characters wide,
// code stepping out wider than the text, and the author at the end.

export const dynamicParams = false;

export function generateStaticParams() {
    return getPosts().map((post) => ({ slug: post.slugs[0]! }));
}

export async function generateMetadata(
    props: PageProps<"/blog/[slug]">,
): Promise<Metadata> {
    const { slug } = await props.params;
    const post = getPost(slug);
    if (!post) notFound();
    const url = `${SITE_URL}${post.url}`;
    const title = smarten(post.data.title);
    const description = smarten(post.data.description ?? "");
    const image = `/og/blog/${slug}/image.png`;
    return {
        title,
        description,
        authors: [{ name: AUTHOR.name, url: AUTHOR.url }],
        alternates: {
            canonical: url,
            types: { "application/rss+xml": `${SITE_URL}/blog/rss.xml` },
        },
        openGraph: {
            type: "article",
            url,
            title,
            description,
            siteName: "Cntrl",
            publishedTime: post.data.date.toISOString(),
            authors: [AUTHOR.url],
            images: [{ url: image, width: 1200, height: 630, alt: title }],
        },
        twitter: {
            card: "summary_large_image",
            creator: `@${AUTHOR.handle}`,
            title,
            description,
            images: [image],
        },
    };
}

export default async function BlogPost(props: PageProps<"/blog/[slug]">) {
    const { slug } = await props.params;
    const post = getPost(slug);
    if (!post) notFound();
    const MDX = post.data.body;
    const title = smarten(post.data.title);
    const description = smarten(post.data.description ?? "");
    const url = `${SITE_URL}${post.url}`;
    const jsonLd = {
        "@context": "https://schema.org",
        "@type": "BlogPosting",
        headline: title,
        description,
        datePublished: post.data.date.toISOString(),
        url,
        mainEntityOfPage: url,
        image: `${SITE_URL}/og/blog/${slug}/image.png`,
        author: { "@type": "Person", name: AUTHOR.name, url: AUTHOR.url },
        publisher: { "@type": "Organization", name: "Cntrl", url: SITE_URL },
    };
    return (
        <main className="w-full flex-1 px-6 pt-14 pb-32 md:pt-24">
            <script
                type="application/ld+json"
                dangerouslySetInnerHTML={{
                    __html: JSON.stringify(jsonLd).replace(/</g, "\\u003c"),
                }}
            />
            <article>
                <header className="mx-auto max-w-[760px] text-center">
                    <Link
                        href="/blog"
                        className="text-fd-muted-foreground hover:text-fd-foreground inline-flex h-8 items-center rounded-md px-2 text-sm transition-colors duration-150"
                    >
                        Blog
                    </Link>
                    <h1 className="font-sg mt-5 text-4xl leading-[1.05] font-semibold tracking-[-0.03em] text-balance md:text-[52px]">
                        {title}
                    </h1>
                    <p className="text-fd-muted-foreground mx-auto mt-5 max-w-[60ch] text-lg text-pretty md:text-xl">
                        {description}
                    </p>
                    <div className="mt-8 flex flex-col items-center gap-1 text-sm">
                        <AuthorLink />
                        <span className="text-fd-muted-foreground tabular-nums">
                            <time dateTime={isoDate(post.data.date)}>
                                {formatDate(post.data.date)}
                            </time>{" "}
                            · {await readingMinutes(post)} min read
                        </span>
                    </div>
                </header>
                <div className="prose [&_figcaption]:text-fd-muted-foreground mx-auto mt-14 max-w-[68ch] md:text-[17px] [&_figcaption]:mt-3 [&_figcaption]:text-center [&_figcaption]:text-sm md:[&_figure]:-mx-12 lg:[&_figure]:-mx-20 [&_figure_img]:my-0 [&_figure_img]:inline-block [&_figure_img]:outline-1 [&_figure_img]:-outline-offset-1 [&_figure_img]:outline-black/10 dark:[&_figure_img]:outline-white/10 [&_figure:has(img)_p]:my-0 [&_figure:has(img)_p]:text-center">
                    <MDX components={getMDXComponents()} />
                </div>
                {/* Its width in the prose's own size, so it lines up with the text. */}
                <footer className="mx-auto mt-20 max-w-[68ch] border-t pt-8 md:text-[17px]">
                    <div className="flex flex-col gap-1 text-sm">
                        <AuthorLink />
                        <p className="text-fd-muted-foreground">{AUTHOR.bio}</p>
                    </div>
                </footer>
            </article>
        </main>
    );
}
