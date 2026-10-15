import { AUTHOR } from "@/lib/blog";
import { cn } from "@/lib/cn";

/** The author's name and X handle as text, linked to the X profile that posts the articles. */
export function AuthorLink({ className }: { className?: string }) {
    return (
        <a
            href={AUTHOR.url}
            target="_blank"
            rel="noreferrer"
            className={cn(
                "text-fd-foreground hover:text-fd-foreground/70 inline-flex items-baseline gap-1.5 font-medium transition-colors duration-150",
                className,
            )}
        >
            {AUTHOR.name}
            <span className="text-fd-muted-foreground font-normal">@{AUTHOR.handle}</span>
        </a>
    );
}
