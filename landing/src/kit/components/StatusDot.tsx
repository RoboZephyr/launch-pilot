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
  scheduled: colors.status.scheduled.fg,
  completed: colors.status.completed.fg,
  stopped: colors.status.stopped.fg,
  error: colors.status.error.fg,
  offline: colors.status.offline.fg,
};

const bgByStatus: Record<ServiceStatus, string> = {
  running: colors.status.running.bg,
  scheduled: colors.status.scheduled.bg,
  completed: colors.status.completed.bg,
  stopped: colors.status.stopped.bg,
  error: colors.status.error.bg,
  offline: colors.status.offline.bg,
};

const borderByStatus: Record<ServiceStatus, string> = {
  running: colors.status.running.border,
  scheduled: colors.status.scheduled.border,
  completed: colors.status.completed.border,
  stopped: colors.status.stopped.border,
  error: colors.status.error.border,
  offline: colors.status.offline.border,
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
