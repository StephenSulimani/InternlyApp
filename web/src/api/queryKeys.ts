import type { JobListParams } from "./jobs";

export const queryKeys = {
  jobs: {
    all: ["jobs"] as const,
    list: (params?: JobListParams) =>
      [...queryKeys.jobs.all, "list", params ?? {}] as const,
    detail: (id: string) => [...queryKeys.jobs.all, "detail", id] as const,
    stats: () => [...queryKeys.jobs.all, "stats"] as const,
    preview: () => [...queryKeys.jobs.all, "preview"] as const,
    locations: () => [...queryKeys.jobs.all, "locations"] as const,
  },
  savedSearches: {
    all: ["saved-searches"] as const,
    list: () => [...queryKeys.savedSearches.all, "list"] as const,
  },
  auth: {
    all: ["auth"] as const,
    session: () => [...queryKeys.auth.all, "session"] as const,
  },
};
