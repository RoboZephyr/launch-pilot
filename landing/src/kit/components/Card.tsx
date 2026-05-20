import { forwardRef, type HTMLAttributes } from "react";
import { cn } from "@/kit/utils/cn";

export interface CardProps extends HTMLAttributes<HTMLDivElement> {
  tone?: "default" | "raised" | "glow";
}

const toneStyles: Record<NonNullable<CardProps["tone"]>, string> = {
  default: "bg-white/[0.04] border border-white/[0.08]",
  raised:
    "bg-[#0B0F1A] border border-white/[0.10] shadow-[0_12px_32px_rgba(0,0,0,0.48)]",
  glow: "bg-white/[0.04] border border-[#22D3EE]/30 shadow-[0_0_0_1px_rgba(34,211,238,0.28),0_16px_48px_rgba(34,211,238,0.16)]",
};

export const Card = forwardRef<HTMLDivElement, CardProps>(
  ({ className, tone = "default", ...props }, ref) => {
    return (
      <div
        ref={ref}
        className={cn(
          "rounded-xl backdrop-blur-sm",
          toneStyles[tone],
          className,
        )}
        {...props}
      />
    );
  },
);

Card.displayName = "Card";
