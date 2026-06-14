// A tiny token store backed by localStorage with an in-memory cache.
//
// Learning note on security: storing JWTs in localStorage is simple and fine
// for a demo, but it is readable by any script on the page, so it is vulnerable
// to XSS. A hardened production setup keeps the refresh token in an httpOnly,
// Secure, SameSite cookie that JavaScript cannot read. The rest of the app is
// written against this module, so swapping the storage strategy is a localized
// change. See docs/architecture.md.

import type { AuthTokens } from "./types";

const KEY = "aura.tokens";
let memory: AuthTokens | null = null;

const isBrowser = () => typeof window !== "undefined";

export function loadTokens(): AuthTokens | null {
  if (memory) return memory;
  if (!isBrowser()) return null;
  try {
    const raw = window.localStorage.getItem(KEY);
    memory = raw ? (JSON.parse(raw) as AuthTokens) : null;
  } catch {
    memory = null;
  }
  return memory;
}

export function saveTokens(tokens: AuthTokens): void {
  memory = tokens;
  if (isBrowser()) {
    window.localStorage.setItem(KEY, JSON.stringify(tokens));
  }
}

export function clearTokens(): void {
  memory = null;
  if (isBrowser()) {
    window.localStorage.removeItem(KEY);
  }
}
