"use client";

import type React from "react";
import { Toaster } from "@/components/ui/toaster";
import { QueryProvider } from "@/providers/query-provider";

export function Providers({ children }: { children: React.ReactNode }) {
  return (
    <QueryProvider>
      {children}
      <Toaster />
    </QueryProvider>
  );
}

