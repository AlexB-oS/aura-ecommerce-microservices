"use client";

// AuthProvider is the single source of truth for "who is signed in" on the
// client. It hydrates the session on mount (by calling /users/me if tokens
// exist) and exposes login / register / logout / updateProfile to the UI.

import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useState,
  type ReactNode,
} from "react";
import * as api from "./api";
import { hasSession } from "./api";
import type { User } from "./types";

type Status = "loading" | "authenticated" | "guest";

interface AuthContextValue {
  user: User | null;
  status: Status;
  login: (email: string, password: string) => Promise<void>;
  register: (input: { email: string; name: string; password: string }) => Promise<void>;
  logout: () => void;
  updateProfile: (name: string) => Promise<void>;
}

const AuthContext = createContext<AuthContextValue | null>(null);

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [status, setStatus] = useState<Status>("loading");

  // On first load, restore the session from a stored token (if any).
  useEffect(() => {
    let active = true;
    if (!hasSession()) {
      setStatus("guest");
      return;
    }
    api
      .fetchMe()
      .then((me) => {
        if (!active) return;
        setUser(me);
        setStatus("authenticated");
      })
      .catch(() => {
        if (!active) return;
        api.logout();
        setStatus("guest");
      });
    return () => {
      active = false;
    };
  }, []);

  const login = useCallback(async (email: string, password: string) => {
    const me = await api.login({ email, password });
    setUser(me);
    setStatus("authenticated");
  }, []);

  const register = useCallback(
    async (input: { email: string; name: string; password: string }) => {
      const me = await api.register(input);
      setUser(me);
      setStatus("authenticated");
    },
    [],
  );

  const logout = useCallback(() => {
    api.logout();
    setUser(null);
    setStatus("guest");
  }, []);

  const updateProfile = useCallback(async (name: string) => {
    const me = await api.updateMe(name);
    setUser(me);
  }, []);

  return (
    <AuthContext.Provider value={{ user, status, login, register, logout, updateProfile }}>
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth(): AuthContextValue {
  const ctx = useContext(AuthContext);
  if (!ctx) {
    throw new Error("useAuth must be used within an <AuthProvider>");
  }
  return ctx;
}
