"use client";

import type React from "react";
import { Toaster } from "@/components/ui/toaster";
import { QueryProvider } from "@/providers/query-provider";
import { useNotifications } from "@/hooks/use-notifications";

export function Providers({ children }: { children: React.ReactNode }) {
  // Автоматическое подключение к SSE для уведомлений
  useNotifications();

  return (
    <QueryProvider>
      {children}
      <Toaster />
    </QueryProvider>
  );
}

