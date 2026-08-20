import { apiClient } from "./client";

export type SavedSearch = {
  id: string;
  name: string;
  q: string;
  type: string;
  location: string;
  recency: "" | "24h" | "3d" | "7d";
  saved: boolean;
  sort: "posted" | "company" | "role" | "location" | "type";
  order: "asc" | "desc";
  created_at?: string;
  updated_at?: string;
};

export type SavedSearchInput = {
  name: string;
  q?: string;
  type?: string;
  location?: string;
  recency?: SavedSearch["recency"];
  saved?: boolean;
  sort?: SavedSearch["sort"];
  order?: SavedSearch["order"];
};

type ApiResponse<T> = {
  success: boolean;
  message: string;
  data?: T;
};

export async function fetchSavedSearches(): Promise<SavedSearch[]> {
  const res = await apiClient<ApiResponse<SavedSearch[]>>("/saved-searches");
  return res.data ?? [];
}

export async function createSavedSearch(
  input: SavedSearchInput,
): Promise<SavedSearch> {
  const res = await apiClient<ApiResponse<SavedSearch>>("/saved-searches", {
    method: "POST",
    body: JSON.stringify(input),
  });
  if (!res.data) {
    throw new Error(res.message || "Failed to save search");
  }
  return res.data;
}

export async function updateSavedSearch(
  id: string,
  input: SavedSearchInput,
): Promise<SavedSearch> {
  const res = await apiClient<ApiResponse<SavedSearch>>(
    `/saved-searches/${id}`,
    {
      method: "PUT",
      body: JSON.stringify(input),
    },
  );
  if (!res.data) {
    throw new Error(res.message || "Failed to update saved search");
  }
  return res.data;
}

export async function deleteSavedSearch(id: string): Promise<void> {
  await apiClient<ApiResponse<unknown>>(`/saved-searches/${id}`, {
    method: "DELETE",
  });
}
