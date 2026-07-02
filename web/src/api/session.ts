import type { UserProfile } from "./auth";

const TOKEN_KEY = "internly_token";
const USER_KEY = "internly_user";

export function saveSession(token: string, user: UserProfile) {
  localStorage.setItem(TOKEN_KEY, token);
  localStorage.setItem(USER_KEY, JSON.stringify(user));
}

export function getToken(): string | null {
  return localStorage.getItem(TOKEN_KEY);
}

export function getStoredUser(): UserProfile | null {
  const raw = localStorage.getItem(USER_KEY);
  if (!raw) {
    return null;
  }

  try {
    return JSON.parse(raw) as UserProfile;
  } catch {
    return null;
  }
}

export function clearSession() {
  localStorage.removeItem(TOKEN_KEY);
  localStorage.removeItem(USER_KEY);
}

export function loadSession(): { token: string; user: UserProfile } | null {
  const token = getToken();
  const user = getStoredUser();

  if (!token || !user) {
    clearSession();
    return null;
  }

  if (isTokenExpired(token)) {
    clearSession();
    return null;
  }

  return { token, user };
}

function isTokenExpired(token: string): boolean {
  const payload = parseJwtPayload(token);
  if (!payload?.exp) {
    return false;
  }

  return payload.exp * 1000 <= Date.now();
}

function parseJwtPayload(token: string): { exp?: number } | null {
  const segment = token.split(".")[1];
  if (!segment) {
    return null;
  }

  try {
    const normalized = segment.replace(/-/g, "+").replace(/_/g, "/");
    return JSON.parse(atob(normalized)) as { exp?: number };
  } catch {
    return null;
  }
}
