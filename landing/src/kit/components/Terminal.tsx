import { type HTMLAttributes, type ReactNode } from "react";
import { cn } from "@/kit/utils/cn";

export interface TerminalProps extends HTMLAttributes<HTMLDivElement> {
  title?: string;
  prompt?: string;
  children?: ReactNode;
}

export function Terminal({
  title = "~ launch-pilot",
  prompt,
  children,
  className,
  ...props
}: TerminalProps) {
  return (
    <div
      className={cn(
        "overflow-hidden rounded-xl border border-white/10 bg-[#03040A] shadow-[0_24px_64px_rgba(0,0,0,0.55)]",
        className,
      )}
      {...props}
    >
      <header className="flex items-center gap-2 border-b border-white/[0.06] bg-white/[0.02] px-4 py-2.5">
        <span className="flex gap-1.5">
          <span className="h-2.5 w-2.5 rounded-full bg-[#FF5F57]" />
          <span className="h-2.5 w-2.5 rounded-full bg-[#FEBC2E]" />
          <span className="h-2.5 w-2.5 rounded-full bg-[#28C840]" />
        </span>
        <span className="ml-2 text-xs font-medium text-white/50 tracking-wide font-mono">
          {title}
        </span>
      </header>
      <pre className="px-4 py-4 text-[13px] leading-relaxed font-mono text-white/85 whitespace-pre-wrap break-words">
        {prompt && (
          <span className="block text-[#22D3EE]">
            <span className="text-white/40">$</span> {prompt}
          </span>
        )}
        {children}
      </pre>
    </div>
  );
}
