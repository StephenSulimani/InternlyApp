import { keepPreviousData, useQuery } from "@tanstack/react-query";
import {
  fetchJob,
  fetchJobLocations,
  fetchJobs,
  type JobListParams,
} from "../api/jobs";
import { queryKeys } from "../api/queryKeys";

/** How often the signed-in board refetches while the page is open. */
export const BOARD_REFETCH_INTERVAL_MS = 60_000;

export function useJobs(params?: JobListParams, enabled = true) {
  return useQuery({
    queryKey: queryKeys.jobs.list(params),
    queryFn: () => fetchJobs(params),
    enabled,
    placeholderData: keepPreviousData,
    staleTime: 30_000,
    refetchInterval: enabled ? BOARD_REFETCH_INTERVAL_MS : false,
    refetchIntervalInBackground: false,
  });
}

export function useJobLocations(enabled = true) {
  return useQuery({
    queryKey: queryKeys.jobs.locations(),
    queryFn: fetchJobLocations,
    enabled,
    staleTime: 60_000,
  });
}

export function useJob(id: string, enabled = true) {
  return useQuery({
    queryKey: queryKeys.jobs.detail(id),
    queryFn: () => fetchJob(id),
    enabled: enabled && Boolean(id),
  });
}
