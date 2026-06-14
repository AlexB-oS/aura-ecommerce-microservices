// API client for the AURA platform.
//
// All calls go to the gateway under /api/v1. Because the browser, the gateway,
// and the API share an origin (everything is served from http://localhost:8080),
// these are relative URLs and there is no browser CORS to deal with.
//
// When running `next dev` standalone (port 3000, no gateway) set
// NEXT_PUBLIC_API_BASE=http://localhost:8080 so calls reach the gateway.

import type { AuthResponse, AuthTokens, ApiErrorBody, User } from "./types";
import { clearTokens, loadTokens, saveTokens } from "./token-store";

const API_BASE = process.env.NEXT_PUBLIC_API_BASE ?? "";
const PREFIX = "/api/v1";

/** Typed error thrown for any non-2xx response, carrying the field map so forms
 *  can highlight individual inputs. */
export class ApiError extends Error {
  status: number;
  code: string;
  fields?: Record<string, string>;

  constructor(status: number, body?: Partial<ApiErrorBody>) {
    super(body?.message ?? "Request failed");
    this.status = status;
    this.code = body?.code ?? "error";
    this.fields = body?.fields;
  }
}

async function rawRequest<T>(path: string, init: RequestInit = {}): Promise<T> {
  const res = await fetch(`${API_BASE}${PREFIX}${path}`, {
    ...init,
    headers: {
      "Content-Type": "application/json",
      ...(init.headers ?? {}),
    },
  });

  const text = await res.text();
  const json = text ? JSON.parse(text) : {};

  if (!res.ok) {
    throw new ApiError(res.status, json?.error);
  }
  return json.data as T;
}

/** Attempt to swap a refresh token for a fresh token pair. Returns the new
 *  tokens or null if the refresh token is no longer valid. */
async function tryRefresh(refreshToken: string): Promise<AuthTokens | null> {
  try {
    const data = await rawRequest<AuthResponse>("/auth/refresh", {
      method: "POST",
      body: JSON.stringify({ refreshToken }),
    });
    saveTokens(data);
    return data;
  } catch {
    clearTokens();
    return null;
  }
}

/** Authenticated request: attaches the access token and, on a 401, transparently
 *  refreshes once and retries before giving up. */
async function authRequest<T>(path: string, init: RequestInit = {}): Promise<T> {
  const tokens = loadTokens();
  const withAuth = (token?: string): RequestInit => ({
    ...init,
    headers: {
      ...(init.headers ?? {}),
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
    },
  });

  try {
    return await rawRequest<T>(path, withAuth(tokens?.accessToken));
  } catch (err) {
    if (err instanceof ApiError && err.status === 401 && tokens?.refreshToken) {
      const refreshed = await tryRefresh(tokens.refreshToken);
      if (refreshed) {
        return await rawRequest<T>(path, withAuth(refreshed.accessToken));
      }
    }
    throw err;
  }
}

/** authed makes an authenticated request to any platform endpoint under
 *  /api/v1, attaching the access token and auto-refreshing on a 401. Other API
 *  modules (cart, orders) build on this. */
export function authed<T>(path: string, init: RequestInit = {}): Promise<T> {
  return authRequest<T>(path, init);
}

// --- Public API ------------------------------------------------------------

export async function register(input: {
  email: string;
  name: string;
  password: string;
}): Promise<User> {
  const data = await rawRequest<AuthResponse>("/auth/register", {
    method: "POST",
    body: JSON.stringify(input),
  });
  saveTokens(data);
  return data.user!;
}

export async function login(input: {
  email: string;
  password: string;
}): Promise<User> {
  const data = await rawRequest<AuthResponse>("/auth/login", {
    method: "POST",
    body: JSON.stringify(input),
  });
  saveTokens(data);
  return data.user!;
}

export function fetchMe(): Promise<User> {
  return authRequest<User>("/users/me");
}

export function updateMe(name: string): Promise<User> {
  return authRequest<User>("/users/me", {
    method: "PATCH",
    body: JSON.stringify({ name }),
  });
}

export function logout(): void {
  clearTokens();
}

export function hasSession(): boolean {
  return loadTokens() !== null;
}
