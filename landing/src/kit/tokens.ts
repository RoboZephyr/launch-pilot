/**
 * Launch Pilot — visual tokens.
 *
 * Aesthetic: macOS-native dev console. Deep graphite background, signal-cyan
 * primary (SSE stream feel), violet accent (developer tool signature). Status
 * semantics mirror launchd service states.
 */

export const colors = {
  bg: {
    base: "#05070D",
    raised: "#0B0F1A",
    sunken: "#03040A",
    surface: "rgba(255, 255, 255, 0.04)",
    surfaceHover: "rgba(255, 255, 255, 0.07)",
  },
  fg: {
    primary: "#F4F6FB",
    muted: "rgba(244, 246, 251, 0.62)",
    subtle: "rgba(244, 246, 251, 0.38)",
    inverse: "#05070D",
  },
  border: {
    subtle: "rgba(255, 255, 255, 0.06)",
    default: "rgba(255, 255, 255, 0.10)",
    strong: "rgba(255, 255, 255, 0.18)",
  },
  brand: {
    primary: "#22D3EE",
    primaryHover: "#67E8F9",
    primaryMuted: "rgba(34, 211, 238, 0.14)",
    accent: "#7C5CFF",
    accentHover: "#A78BFA",
    accentMuted: "rgba(124, 92, 255, 0.16)",
  },
  status: {
    running: {
      fg: "#34D399",
      bg: "rgba(52, 211, 153, 0.12)",
      border: "rgba(52, 211, 153, 0.36)",
      label: "Running",
    },
    idle: {
      fg: "#94A3B8",
      bg: "rgba(148, 163, 184, 0.12)",
      border: "rgba(148, 163, 184, 0.32)",
      label: "Idle",
    },
    loaded: {
      fg: "#22D3EE",
      bg: "rgba(34, 211, 238, 0.12)",
      border: "rgba(34, 211, 238, 0.36)",
      label: "Loaded",
    },
    warning: {
      fg: "#FBBF24",
      bg: "rgba(251, 191, 36, 0.12)",
      border: "rgba(251, 191, 36, 0.38)",
      label: "Degraded",
    },
    failed: {
      fg: "#F87171",
      bg: "rgba(248, 113, 113, 0.12)",
      border: "rgba(248, 113, 113, 0.40)",
      label: "Failed",
    },
    unknown: {
      fg: "#64748B",
      bg: "rgba(100, 116, 139, 0.12)",
      border: "rgba(100, 116, 139, 0.28)",
      label: "Unknown",
    },
  },
} as const;

export type ServiceStatus = keyof typeof colors.status;

export const SERVICE_STATUS_KEYS: readonly ServiceStatus[] = [
  "running",
  "idle",
  "loaded",
  "warning",
  "failed",
  "unknown",
] as const;

export const typography = {
  fontFamily: {
    sans: "var(--font-space-grotesk), ui-sans-serif, system-ui, -apple-system, sans-serif",
    mono: "ui-monospace, SFMono-Regular, 'SF Mono', Menlo, Consolas, monospace",
    display:
      "var(--font-space-grotesk), ui-sans-serif, system-ui, -apple-system, sans-serif",
  },
  fontSize: {
    xs: "0.75rem",
    sm: "0.875rem",
    base: "1rem",
    lg: "1.125rem",
    xl: "1.25rem",
    "2xl": "1.5rem",
    "3xl": "1.875rem",
    "4xl": "2.25rem",
    "5xl": "3rem",
    "6xl": "3.75rem",
    "7xl": "4.5rem",
  },
  fontWeight: {
    regular: 400,
    medium: 500,
    semibold: 600,
    bold: 700,
  },
  lineHeight: {
    tight: 1.1,
    snug: 1.25,
    normal: 1.5,
    relaxed: 1.65,
  },
  letterSpacing: {
    tight: "-0.02em",
    normal: "0",
    wide: "0.04em",
    wider: "0.08em",
  },
} as const;

export const spacing = {
  px: "1px",
  "0": "0",
  "1": "0.25rem",
  "2": "0.5rem",
  "3": "0.75rem",
  "4": "1rem",
  "5": "1.25rem",
  "6": "1.5rem",
  "8": "2rem",
  "10": "2.5rem",
  "12": "3rem",
  "16": "4rem",
  "20": "5rem",
  "24": "6rem",
  "32": "8rem",
} as const;

export const radii = {
  none: "0",
  sm: "4px",
  md: "8px",
  lg: "12px",
  xl: "16px",
  "2xl": "24px",
  pill: "9999px",
} as const;

export const shadow = {
  card: "0 1px 0 rgba(255,255,255,0.04) inset, 0 12px 32px rgba(0,0,0,0.48)",
  glow: "0 0 0 1px rgba(34,211,238,0.32), 0 12px 48px rgba(34,211,238,0.18)",
  ring: "0 0 0 1px rgba(255,255,255,0.12)",
  floating: "0 24px 64px rgba(0,0,0,0.55)",
} as const;

export const animation = {
  duration: {
    instant: "80ms",
    fast: "160ms",
    base: "240ms",
    slow: "400ms",
    slower: "640ms",
  },
  easing: {
    standard: "cubic-bezier(0.4, 0, 0.2, 1)",
    emphasized: "cubic-bezier(0.2, 0, 0, 1)",
    exit: "cubic-bezier(0.4, 0, 1, 1)",
    spring: "cubic-bezier(0.22, 1, 0.36, 1)",
  },
  pulse: {
    duration: "1600ms",
    easing: "cubic-bezier(0.4, 0, 0.6, 1)",
  },
} as const;

export const breakpoints = {
  sm: "640px",
  md: "768px",
  lg: "1024px",
  xl: "1280px",
  "2xl": "1536px",
} as const;

export const tokens = {
  colors,
  typography,
  spacing,
  radii,
  shadow,
  animation,
  breakpoints,
} as const;

export type Tokens = typeof tokens;
