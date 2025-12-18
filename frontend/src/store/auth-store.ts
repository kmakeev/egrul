import { create } from "zustand";
import { persist } from "zustand/middleware";
import type { AuthState, User } from "@/types";

interface AuthStore extends AuthState {
  token: string | null;
  setUser: (user: User | null) => void;
  setToken: (token: string | null) => void;
  logout: () => void;
}

export const useAuthStore = create<AuthStore>()(
  persist(
    (set) => ({
      user: null,
      isAuthenticated: false,
      isLoading: false,
      token: null,
      setUser: (user) =>
        set({
          user,
          isAuthenticated: !!user,
        }),
      setToken: (token) =>
        set({
          token,
          isAuthenticated: token ? true : false,
        }),
      logout: () =>
        set({
          user: null,
          token: null,
          isAuthenticated: false,
        }),
    }),
    {
      name: "auth-storage",
    }
  )
);

