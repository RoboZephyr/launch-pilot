import { type HTMLAttributes } from "react";
import { cn } from "@/kit/utils/cn";

type BadgeVariant = "neutral" | "brand" | "accent";

export interface BadgeProps extends HTMLAttributes<HTMLSpanElement> {
  variant?: BadgeVariant;
}

const variantStyles: Record<BadgeVariant, string> = {
  neutral: "bg-white/[0.05] text-white/70 border-white/10",
  brand: "bg-[#22D3EE]/10 text-[#67E8F9] border-[#22D3EE]/30",
  accent: "bg-[#7C5CFF]/10 text-[#A78BFA] border-[#7C5CFF]/30",
};

export function Badge({
  variant = "neutral",
  className,
  ...props
}: BadgeProps) {
  return (
    <span
      className={cn(
        "inline-flex items-center gap-1.5 rounded-full border px-2.5 py-0.5 text-[11px] font-medium uppercase tracking-wider",
        variantStyles[variant],
        className,
      )}
      {...props}
    />
  );
}
