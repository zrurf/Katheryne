import { cn } from "../../lib/utils";
import { getServerApiBase } from "../../services/config";

interface AvatarProps {
  src?: string;
  name?: string;
  size?: "sm" | "md" | "lg" | "xl";
  online?: boolean;
  class?: string;
}

/** Resolve a relative OSS URL to a full URL using the configured server host. */
function resolveUrl(url?: string): string | undefined {
  if (!url) return undefined;
  if (url.startsWith("http://") || url.startsWith("https://")) return url;
  return getServerApiBase() + url;
}

export function Avatar(props: AvatarProps) {
  const sizeClasses = {
    sm: "w-8 h-8 text-xs",
    md: "w-10 h-10 text-sm",
    lg: "w-12 h-12 text-base",
    xl: "w-16 h-16 text-xl",
  };

  const onlineSize = {
    sm: "w-2 h-2 right-0 bottom-0",
    md: "w-2.5 h-2.5 right-0 bottom-0",
    lg: "w-3 h-3 right-0 bottom-0",
    xl: "w-3.5 h-3.5 right-0.5 bottom-0.5",
  };

  const initials = (props.name || "?")
    .split(" ")
    .map((n) => n[0])
    .join("")
    .toUpperCase()
    .slice(0, 2);

  const size = props.size || "md";

  return (
    <div class={cn("relative shrink-0", props.class)}>
      {props.src ? (
        <img
          src={resolveUrl(props.src)}
          alt={props.name || ""}
          class={cn("rounded-full object-cover", sizeClasses[size])}
        />
      ) : (
        <div
          class={cn(
            "rounded-full bg-primary/20 text-primary flex items-center justify-center font-semibold",
            sizeClasses[size]
          )}
        >
          {initials}
        </div>
      )}
      {props.online !== undefined && (
        <span
          class={cn(
            "absolute rounded-full border-2 border-bg",
            onlineSize[size],
            props.online ? "bg-success" : "bg-text-muted"
          )}
        />
      )}
    </div>
  );
}