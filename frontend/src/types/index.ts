export * from "../lib/api";

export interface User {
  id: string;
  email: string;
  firstName: string;
  lastName: string;
  createdAt: string;
}

export interface AuthState {
  user: User | null;
  isAuthenticated: boolean;
  isLoading: boolean;
}

export interface SearchFilters {
  query: string;
  entityType?: "all" | "legal" | "entrepreneur";
  region?: string;
  status?: string;
  dateFrom?: string;
  dateTo?: string;
}

export interface SearchHistoryItem {
  id: string;
  query: string;
  createdAt: string;
  filters: SearchFilters;
}

export interface WatchlistItem {
  id: string;
  ogrn?: string;
  ogrnip?: string;
  name: string;
  type: "legal" | "entrepreneur";
  addedAt: string;
}

