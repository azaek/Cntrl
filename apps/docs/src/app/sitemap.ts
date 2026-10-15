import { getPosts } from "@/lib/blog";
import { source } from "@/lib/source";
import type { MetadataRoute } from "next";

const BASE_URL = "https://www.cntrl.pw";

export default function sitemap(): MetadataRoute.Sitemap {
    const docs = source.getPages().map((page) => ({
        url: `${BASE_URL}${page.url}`,
        lastModified: new Date(),
        changeFrequency: "weekly" as const,
        priority: 0.7,
    }));

    // The blog once it has a post (D118), each post dated by its own date.
    const posts = getPosts();
    const blog =
        posts.length === 0
            ? []
            : [
                  {
                      url: `${BASE_URL}/blog`,
                      lastModified: posts[0]!.data.date,
                      changeFrequency: "weekly" as const,
                      priority: 0.8,
                  },
                  ...posts.map((post) => ({
                      url: `${BASE_URL}${post.url}`,
                      lastModified: post.data.date,
                      changeFrequency: "monthly" as const,
                      priority: 0.7,
                  })),
              ];

    return [
        {
            url: BASE_URL,
            lastModified: new Date(),
            changeFrequency: "monthly",
            priority: 1.0,
        },
        {
            url: `${BASE_URL}/docs`,
            lastModified: new Date(),
            changeFrequency: "weekly",
            priority: 0.9,
        },
        ...docs,
        ...blog,
    ];
}
