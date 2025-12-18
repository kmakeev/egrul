"use client";

import type { ReactNode } from "react";
import { useState } from "react";
import {
  QueryClient,
  QueryClientProvider,
  type DefaultOptions,
} from "@tanstack/react-query";

const defaultOptions: DefaultOptions = {
  queries: {
    staleTime: 60 * 1000, // 1 минута
    gcTime: 5 * 60 * 1000, // 5 минут
    refetchOnWindowFocus: false,
    refetchOnReconnect: true,
    retry: 2,
  },
  mutations: {
    retry: 1,
  },
};

export function createQueryClient() {
  return new QueryClient({
    defaultOptions,
  });
}

export function QueryProvider({ children }: { children: ReactNode }) {
  const [queryClient] = useState(() => createQueryClient());

  return <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>;
}
