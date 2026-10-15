import { AUTHOR } from "@/lib/blog";
import { cn } from "@/lib/cn";

/** Simple Icons' X mark. */
export function XMark({ className }: { className?: string }) {
    return (
        <svg
            viewBox="0 0 24 24"
            aria-hidden="true"
            fill="currentColor"
            className={cn("size-3", className)}
        >
            <path d="M14.234 10.162 22.977 0h-2.072l-7.591 8.824L7.251 0H.258l9.168 13.343L.258 24H2.33l8.016-9.318L16.749 24h6.993zm-2.837 3.299-.929-1.329L3.076 1.56h3.182l5.965 8.532.929 1.329 7.754 11.09h-3.182z" />
        </svg>
    );
}

export function Avatar({ size = 28, className }: { size?: number; className?: string }) {
    return (
        // A remote avatar at its own size; next/image would only add a hop.
        // eslint-disable-next-line @next/next/no-img-element
        <img
            src={AUTHOR.avatar}
            alt=""
            width={size}
            height={size}
            className={cn(
                "shrink-0 rounded-full outline-1 -outline-offset-1 outline-black/10 dark:outline-white/10",
                className,
            )}
        />
    );
}

/** The author's name and X handle, linked to the X profile that posts the articles. */
export function AuthorLink({ className }: { className?: string }) {
    return (
        <a
            href={AUTHOR.url}
            target="_blank"
            rel="noreferrer"
            aria-label={`${AUTHOR.name} on X`}
            className={cn(
                "text-fd-foreground hover:text-fd-foreground/70 inline-flex items-center gap-1.5 font-medium transition-colors duration-150",
                className,
            )}
        >
            {AUTHOR.name}
            <span className="text-fd-muted-foreground inline-flex items-center gap-1 font-normal">
                <XMark />@{AUTHOR.handle}
            </span>
        </a>
    );
}
