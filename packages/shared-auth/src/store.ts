import { create } from "zustand";
import { persist } from "zustand/middleware";
import Cookies from "js-cookie";

export interface User {
  id: string; email: string; username: string; display_name: string;
  phone: string; avatar_url: string; status: string; created_at: string;
}

export interface AuthTokens {
  access_token: string; refresh_token: string; expires_in: number; token_type: string;
}

interface AuthState {
  user: User | null;
  tokens: AuthTokens | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  setUser: (user: User | null) => void;
  setTokens: (tokens: AuthTokens | null) => void;
  login: (user: User, tokens: AuthTokens) => void;
  logout: () => void;
  refreshAccessToken: () => Promise<boolean>;
}

const ACCESS_TOKEN_KEY = "access_token";
const REFRESH_TOKEN_KEY = "refresh_token";

export const useAuthStore = create<AuthState>()(
  persist(
    (set, get) => ({
      user: null,
      tokens: null,
      isAuthenticated: false,
      isLoading: false,

      setUser: (user) => set({ user, isAuthenticated: !!user }),
      setTokens: (tokens) => {
        set({ tokens, isAuthenticated: !!tokens });
        if (tokens) {
          Cookies.set(ACCESS_TOKEN_KEY, tokens.access_token, { expires: tokens.expires_in / 86400, sameSite: "strict" });
          Cookies.set(REFRESH_TOKEN_KEY, tokens.refresh_token, { expires: 7, sameSite: "strict" });
        }
      },

      login: (user, tokens) => {
        set({ user, tokens, isAuthenticated: true });
        Cookies.set(ACCESS_TOKEN_KEY, tokens.access_token, { expires: tokens.expires_in / 86400, sameSite: "strict" });
        Cookies.set(REFRESH_TOKEN_KEY, tokens.refresh_token, { expires: 7, sameSite: "strict" });
      },

      logout: () => {
        set({ user: null, tokens: null, isAuthenticated: false });
        Cookies.remove(ACCESS_TOKEN_KEY);
        Cookies.remove(REFRESH_TOKEN_KEY);
      },

      refreshAccessToken: async () => {
        const tokens = get().tokens;
        if (!tokens?.refresh_token) return false;
        try {
          const res = await fetch("/api/auth/refresh", {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({ refresh_token: tokens.refresh_token }),
          });
          if (!res.ok) { get().logout(); return false; }
          const data = await res.json();
          const newTokens = data.data || data;
          get().setTokens(newTokens);
          return true;
        } catch {
          get().logout();
          return false;
        }
      },
    }),
    { name: "tiki-auth" }
  )
);
