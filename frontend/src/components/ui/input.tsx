import * as React from "react";
import { cn } from "@/lib/utils";

const Input = React.forwardRef<
  HTMLInputElement,
  React.InputHTMLAttributes<HTMLInputElement>
>(({ className, ...props }, ref) => (
  <input
    ref={ref}
    className={cn(
      "flex h-9 w-full rounded-md border border-[var(--color-surface0)] bg-[var(--color-mantle)] px-3 py-1 text-sm text-[var(--color-text)] placeholder:text-[var(--color-overlay0)] outline-none transition-colors focus:border-[var(--color-mauve)] disabled:opacity-50",
      className
    )}
    {...props}
  />
));
Input.displayName = "Input";

export { Input };
