import { create } from "zustand";
import { persist } from "zustand/middleware";
import type { SearchFilters, SearchHistoryItem } from "@/types";

interface SearchStore {
  filters: SearchFilters;
  history: SearchHistoryItem[];
  setFilters: (filters: Partial<SearchFilters>) => void;
  resetFilters: () => void;
  addToHistory: (filters: SearchFilters) => void;
  clearHistory: () => void;
}

const defaultFilters: SearchFilters = {
  query: "",
  entityType: "all",
};

export const useSearchStore = create<SearchStore>()(
  persist(
    (set, get) => ({
      filters: defaultFilters,
      history: [],
      setFilters: (newFilters) =>
        set((state) => ({
          filters: { ...state.filters, ...newFilters },
        })),
      resetFilters: () => set({ filters: defaultFilters }),
      addToHistory: (filters) => {
        const item: SearchHistoryItem = {
          id: `${Date.now()}-${Math.random().toString(36).slice(2, 8)}`,
          query: filters.query,
          createdAt: new Date().toISOString(),
          filters,
        };
        const current = get().history;
        set({ history: [item, ...current].slice(0, 50) });
      },
      clearHistory: () => set({ history: [] }),
    }),
    {
      name: "search-storage",
    }
  )
);

