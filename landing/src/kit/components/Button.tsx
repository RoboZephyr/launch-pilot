"use client";

import { forwardRef, type ButtonHTMLAttributes } from "react";
import { cn } from "@/kit/utils/cn";

type ButtonVariant = "primary" | "secondary" | "ghost" | "outline";
type ButtonSize = "sm" | "md" | "lg";

export interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: ButtonVariant;
  size?: ButtonSize;
}

const variantStyles: Record<ButtonVariant, string> = {
  primary:
    "bg-[#22D3EE] text-[#05070D] hover:bg-[#67E8F9] shadow-[0_0_0_1px_rgba(34,211,238,0.4),0_8px_32px_rgba(34,211,238,0.28)]",
  secondary:
    "bg-white/5 text-foreground hover:bg-white/10 border border-white/10",
  ghost: "bg-transparent text-foreground/80 hover:bg-white/5 hover:text-foreground",
  outline:
    "bg-transparent text-foreground border border-white/20 hover:bg-white/5 hover:border-white/40",
};

const sizeStyles: Record<ButtonSize, string> = {
  sm: "h-8 px-3 text-sm rounded-md",
  md: "h-10 px-4 text-sm rounded-lg",
  lg: "h-12 px-6 text-base rounded-lg",
};

export const Button = forwardRef<HTMLButtonElement, ButtonProps>(
  ({ className, variant = "primary", size = "md", ...props }, ref) => {
    return (
      <button
        ref={ref}
        className={cn(
          "inline-flex items-center justify-center gap-2 font-medium tracking-tight",
          "transition-colors duration-200 ease-out",
          "focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-[#22D3EE]/60 focus-visible:ring-offset-2 focus-visible:ring-offset-background",
          "disabled:opacity-50 disabled:pointer-events-none",
          variantStyles[variant],
          sizeStyles[size],
          className,
        )}
        {...props}
      />
    );
  },
);

Button.displayName = "Button";
