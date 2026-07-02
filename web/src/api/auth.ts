import { apiClient } from "./client";

export type UserProfile = {
  id: string;
  first_name: string;
  last_name: string;
  email: string;
  is_admin: boolean;
  is_active: boolean;
  is_premium: boolean;
};

export type LoginResponse = {
  success: boolean;
  message: string;
  token: string;
  user: UserProfile;
};

export type LoginInput = {
  email: string;
  password: string;
};

const TOKEN_KEY = "internly_token";

export async function login(input: LoginInput): Promise<LoginResponse> {
  return apiClient<LoginResponse>("/login", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(input),
  });
}

export function saveToken(token: string) {
  localStorage.setItem(TOKEN_KEY, token);
}

export function getToken(): string | null {
  return localStorage.getItem(TOKEN_KEY);
}

export function clearToken() {
  localStorage.removeItem(TOKEN_KEY);
}
