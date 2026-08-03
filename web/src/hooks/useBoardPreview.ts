import { useQuery } from "@tanstack/react-query";
import { fetchBoardPreview } from "../api/jobs";
import { queryKeys } from "../api/queryKeys";
import { jobToListing } from "../lib/listings";

/** Public homepage board preview (5 recent jobs). */
export function useBoardPreview() {
  return useQuery({
    queryKey: queryKeys.jobs.preview(),
    queryFn: fetchBoardPreview,
    staleTime: 60_000,
    select: (jobs) => jobs.map(jobToListing),
  });
}
