import { type HTMLAttributes } from "react";
import { cn } from "@/kit/utils/cn";
import { colors, type ServiceStatus } from "@/kit/tokens";

export interface StatusDotProps extends HTMLAttributes<HTMLSpanElement> {
  status: ServiceStatus;
  pulse?: boolean;
  label?: boolean;
}

const dotColorByStatus: Record<ServiceStatus, string> = {
  running: colors.status.running.fg,
  idle: colors.status.idle.fg,
  loaded: colors.status.loaded.fg,
  warning: colors.status.warning.fg,
  failed: colors.status.failed.fg,
  unknown: colors.status.unknown.fg,
};

const bgByStatus: Record<ServiceStatus, string> = {
  running: colors.status.running.bg,
  idle: colors.status.idle.bg,
  loaded: colors.status.loaded.bg,
  warning: colors.status.warning.bg,
  failed: colors.status.failed.bg,
  unknown: colors.status.unknown.bg,
};

const borderByStatus: Record<ServiceStatus, string> = {
  running: colors.status.running.border,
  idle: colors.status.idle.border,
  loaded: colors.status.loaded.border,
  warning: colors.status.warning.border,
  failed: colors.status.failed.border,
  unknown: colors.status.unknown.border,
};

export function StatusDot({
  status,
  pulse = false,
  label = false,
  className,
  ...props
}: StatusDotProps) {
  const color = dotColorByStatus[status];
  const statusMeta = colors.status[status];

  if (!label) {
    return (
      <span
        role="status"
        aria-label={statusMeta.label}
        className={cn("relative inline-flex h-2 w-2", className)}
        {...props}
      >
        {pulse && (
          <span
            className="absolute inline-flex h-full w-full rounded-full opacity-60 animate-ping"
            style={{ backgroundColor: color }}
          />
        )}
        <span
          className="relative inline-flex h-2 w-2 rounded-full"
          style={{ backgroundColor: color, boxShadow: `0 0 8px ${color}` }}
        />
      </span>
    );
  }

  return (
    <span
      role="status"
      aria-label={statusMeta.label}
      className={cn(
        "inline-flex items-center gap-2 rounded-full border px-2.5 py-1 text-xs font-medium tracking-wide",
        className,
      )}
      style={{
        color,
        backgroundColor: bgByStatus[status],
        borderColor: borderByStatus[status],
      }}
      {...props}
    >
      <span className="relative inline-flex h-1.5 w-1.5">
        {pulse && (
          <span
            className="absolute inline-flex h-full w-full rounded-full opacity-70 animate-ping"
            style={{ backgroundColor: color }}
          />
        )}
        <span
          className="relative inline-flex h-1.5 w-1.5 rounded-full"
          style={{ backgroundColor: color }}
        />
      </span>
      {statusMeta.label}
    </span>
  );
}
