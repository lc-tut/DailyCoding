import * as React from "react";
import { cva, type VariantProps } from "class-variance-authority";
import { cn } from "@/lib/utils";

const badgeVariants = cva(
  "inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-semibold transition-colors",
  {
    variants: {
      variant: {
        default:
          "bg-[var(--color-mauve)] text-[var(--color-mantle)]",
        secondary:
          "bg-[var(--color-surface0)] text-[var(--color-subtext0)]",
        success:
          "bg-[color-mix(in_srgb,var(--color-green)_20%,transparent)] border border-[var(--color-green)] text-[var(--color-green)]",
        destructive:
          "bg-[color-mix(in_srgb,var(--color-red)_20%,transparent)] border border-[var(--color-red)] text-[var(--color-red)]",
        outline:
          "border border-[var(--color-surface0)] text-[var(--color-subtext0)]",
      },
    },
    defaultVariants: { variant: "default" },
  }
);

export interface BadgeProps
  extends React.HTMLAttributes<HTMLDivElement>,
    VariantProps<typeof badgeVariants> {}

function Badge({ className, variant, ...props }: BadgeProps) {
  return (
    <div className={cn(badgeVariants({ variant }), className)} {...props} />
  );
}

export { Badge, badgeVariants };
