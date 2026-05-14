import { cn } from "../../lib/utils";
import { splitProps, type ComponentProps, type ValidComponent } from "solid-js";
import { Dynamic } from "solid-js/web";

interface ButtonProps extends ComponentProps<"button"> {
  variant?: "primary" | "secondary" | "ghost" | "danger";
  size?: "sm" | "md" | "lg";
  as?: ValidComponent;
}

export function Button(props: ButtonProps) {
  const [local, rest] = splitProps(props, ["variant", "size", "class", "as", "children"]);

  const variantClasses = {
    primary:
      "bg-primary hover:bg-primary-dark text-white shadow-lg shadow-primary/25",
    secondary:
      "bg-surface hover:bg-surface-hover text-text border border-border",
    ghost:
      "hover:bg-surface-hover text-text-secondary hover:text-text",
    danger:
      "bg-danger/10 hover:bg-danger/20 text-danger border border-danger/20",
  };

  const sizeClasses = {
    sm: "px-3 py-1.5 text-xs rounded-lg",
    md: "px-4 py-2 text-sm rounded-xl",
    lg: "px-6 py-3 text-base rounded-xl",
  };

  const Comp = local.as || "button";

  return (
    <Dynamic
      component={Comp}
      class={cn(
        "inline-flex items-center justify-center gap-2 font-medium transition-all duration-200 cursor-pointer disabled:opacity-50 disabled:cursor-not-allowed active:scale-[0.98]",
        variantClasses[local.variant || "primary"],
        sizeClasses[local.size || "md"],
        local.class
      )}
      {...rest}
    >
      {local.children}
    </Dynamic>
  );
}