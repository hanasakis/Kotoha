import { create } from 'zustand';
import type { User } from '../types/auth';
import * as authApi from '../api/auth';

interface AuthState {
  user: User | null;
  accessToken: string | null;
  refreshToken: string | null;
  isAuthenticated: boolean;
  locale: string;

  login: (email: string, password: string) => Promise<void>;
  register: (email: string, password: string, nickname?: string) => Promise<void>;
  logout: () => Promise<void>;
  setTokens: (accessToken: string, refreshToken: string) => void;
  setUser: (user: User) => void;
  setLocale: (locale: string) => void;
  hydrate: () => void;
}

export const useAuthStore = create<AuthState>((set, get) => ({
  user: null,
  accessToken: null,
  refreshToken: null,
  isAuthenticated: false,
  locale: 'zh',

  setTokens: (accessToken: string, refreshToken: string) => {
    localStorage.setItem('kotoha_access_token', accessToken);
    localStorage.setItem('kotoha_refresh_token', refreshToken);
    set({ accessToken, refreshToken, isAuthenticated: true });
  },

  setUser: (user: User) => set({ user }),

  setLocale: (locale: string) => {
    localStorage.setItem('kotoha_locale', locale);
    set({ locale });
  },

  login: async (email: string, password: string) => {
    const result = await authApi.login({ email, password });
    get().setTokens(result.access_token, result.refresh_token);
    set({ user: result.user });
  },

  register: async (email: string, password: string, nickname?: string) => {
    const result = await authApi.register({ email, password, nickname });
    get().setTokens(result.access_token, result.refresh_token);
    set({ user: result.user });
  },

  logout: async () => {
    try {
      const token = get().refreshToken;
      if (token) await authApi.logout(token);
    } catch {
      // Ignore logout errors — clear local state anyway
    }
    localStorage.removeItem('kotoha_access_token');
    localStorage.removeItem('kotoha_refresh_token');
    set({ user: null, accessToken: null, refreshToken: null, isAuthenticated: false });
  },

  hydrate: () => {
    const accessToken = localStorage.getItem('kotoha_access_token');
    const refreshToken = localStorage.getItem('kotoha_refresh_token');
    const locale = localStorage.getItem('kotoha_locale');
    if (accessToken && refreshToken) {
      set({ accessToken, refreshToken, isAuthenticated: true });
    }
    if (locale) {
      set({ locale });
    }
  },
}));
