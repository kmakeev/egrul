import { create } from "zustand";
import { persist, createJSONStorage } from "zustand/middleware";

// ==================== User Store ====================
// Временное решение для хранения email пользователя без полной аутентификации
// В будущем заменить на полноценную систему аутентификации

interface UserState {
  email: string | null;
  setEmail: (email: string) => void;
  clearEmail: () => void;
}

export const useUserStore = create<UserState>()(
  persist(
    (set) => ({
      email: null,
      setEmail: (email: string) => set({ email }),
      clearEmail: () => set({ email: null }),
    }),
    {
      name: "egrul-user-storage",
      storage: createJSONStorage(() => localStorage),
    }
  )
);
