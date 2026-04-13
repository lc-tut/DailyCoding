import * as React from "react";
import { cn } from "@/lib/utils";

const Textarea = React.forwardRef<
  HTMLTextAreaElement,
  React.TextareaHTMLAttributes<HTMLTextAreaElement>
>(({ className, ...props }, ref) => (
  <textarea
    ref={ref}
    className={cn(
      "flex w-full rounded-md border border-[var(--color-surface0)] bg-[var(--color-mantle)] px-3 py-2 text-sm text-[var(--color-text)] placeholder:text-[var(--color-overlay0)] outline-none transition-colors focus:border-[var(--color-mauve)] resize-vertical disabled:opacity-50",
      className
    )}
    {...props}
  />
));
Textarea.displayName = "Textarea";

export { Textarea };
