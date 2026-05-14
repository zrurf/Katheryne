import { cn } from "../../lib/utils";
import { splitProps, type ComponentProps } from "solid-js";

interface InputProps extends ComponentProps<"input"> {
  label?: string;
  error?: string;
  icon?: unknown;
}

export function Input(props: InputProps) {
  const [local, rest] = splitProps(props, ["label", "error", "icon", "class"]);

  return (
    <div class="w-full">
      {local.label && (
        <label class="block text-sm font-medium text-text-secondary mb-1.5">
          {local.label}
        </label>
      )}
      <div class="relative">
        <input
          class={cn(
            "w-full px-4 py-2.5 bg-surface border border-border rounded-xl text-text placeholder:text-text-muted",
            "focus:outline-none focus:border-primary focus:ring-1 focus:ring-primary/30",
            "transition-all duration-200 text-sm",
            local.error && "border-danger focus:border-danger focus:ring-danger/30",
            local.class
          )}
          {...rest}
        />
      </div>
      {local.error && (
        <p class="mt-1 text-xs text-danger">{local.error}</p>
      )}
    </div>
  );
}