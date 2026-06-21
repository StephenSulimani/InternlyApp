export const queryKeys = {
  jobs: {
    all: ["jobs"] as const,
    list: (filters?: { type?: string; location?: string }) =>
      [...queryKeys.jobs.all, "list", filters ?? {}] as const,
    detail: (id: string) => [...queryKeys.jobs.all, "detail", id] as const,
  },
  auth: {
    all: ["auth"] as const,
    session: () => [...queryKeys.auth.all, "session"] as const,
  },
};
