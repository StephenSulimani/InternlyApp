import { useQuery } from "@tanstack/react-query";
import { fetchJobStats } from "../api/jobs";
import { queryKeys } from "../api/queryKeys";

/** Board stats — public. */
export function useJobStats() {
  return useQuery({
    queryKey: queryKeys.jobs.stats(),
    queryFn: fetchJobStats,
    staleTime: 60_000,
  });
}
