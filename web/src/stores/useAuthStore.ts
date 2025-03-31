import { create } from 'zustand';
import { persist } from 'zustand/middleware';

export type User = {
  id: string;
  email: string;
  displayName: string;
};

type AuthState = {
  isAuthenticated: boolean;
  token: string;
  user?: User | null;
  login: (token: string, user: User) => void;
  logout: () => void;
};

export const useAuthStore = create<AuthState>()(
  persist(
    (set) => ({
      isAuthenticated: false,
      token: '',
      user: null,
      login: (token, user) => set({ isAuthenticated: true, token, user }),
      logout: () => set({ isAuthenticated: false, token: '', user: null }),
    }),
    {
      name: 'auth-storage',
    },
  ),
);
