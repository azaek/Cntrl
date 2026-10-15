import {
    applyMdxPreset,
    defineCollections,
    defineConfig,
    defineDocs,
    frontmatterSchema,
    metaSchema,
} from "fumadocs-mdx/config";
import remarkSmartypants from "remark-smartypants";
import { z } from "zod";

// You can customise Zod schemas for frontmatter and `meta.json` here
// see https://fumadocs.dev/docs/mdx/collections
export const docs = defineDocs({
    dir: "content/docs",
    docs: {
        schema: frontmatterSchema.extend({
            badge: z
                .enum(["New", "Updated", "Deprecated", "Beta", "Experimental"])
                .optional(),
        }),
        postprocess: {
            includeProcessedMarkdown: true,
        },
    },
    meta: {
        schema: metaSchema,
    },
});

// The blog at /blog (D118): one article a week, newest first. A draft shows
// only in development, so a deploy of a working tree with one in it stays clean.
export const blog = defineCollections({
    type: "doc",
    dir: "content/blog",
    schema: frontmatterSchema.extend({
        date: z.coerce.date(),
        draft: z.boolean().default(false),
    }),
    postprocess: {
        includeProcessedMarkdown: true,
    },
    // Curly quotes and real dashes in the prose; code keeps what was typed.
    mdxOptions: applyMdxPreset({ remarkPlugins: [remarkSmartypants] }),
});

export default defineConfig({
    mdxOptions: {
        // MDX options
    },
});
