import { useQuery } from "@tanstack/react-query";
import { fetchJob, fetchJobs } from "../api/jobs";
import { queryKeys } from "../api/queryKeys";

type JobListFilters = {
  type?: string;
  location?: string;
};

/**
 * List jobs from the API. Disabled until VITE_ENABLE_API=true.
 * The homepage uses public /board/preview via useBoardPreview.
 */
export function useJobs(filters?: JobListFilters) {
  return useQuery({
    queryKey: queryKeys.jobs.list(filters),
    queryFn: () => fetchJobs(),
    enabled: import.meta.env.VITE_ENABLE_API === "true",
  });
}

export function useJob(id: string) {
  return useQuery({
    queryKey: queryKeys.jobs.detail(id),
    queryFn: () => fetchJob(id),
    enabled: import.meta.env.VITE_ENABLE_API === "true" && Boolean(id),
  });
}
