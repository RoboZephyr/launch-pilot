"use client";

import { ReactNode } from "react";
import { useSmoothScroll } from "@/page/hooks/use-smooth-scroll";

export function Providers({ children }: { children: ReactNode }) {
  useSmoothScroll();
  return <>{children}</>;
}
