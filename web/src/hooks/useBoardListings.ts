import { useEffect, useMemo, useState } from "react";
import { useJobLocations, useJobs } from "./useJobs";
import {
  uniqueSorted,
  type BoardFilters,
  type BoardSort,
} from "../lib/boardFilters";
import { jobToListing } from "../lib/listings";
import type { JobListParams } from "../api/jobs";

export const BOARD_PAGE_SIZE = 50;

const DEFAULT_TYPES = ["Internship", "Full Time"];

export type BoardFacets = {
  types: string[];
  locations: string[];
};

function useDebouncedValue<T>(value: T, delayMs: number): T {
  const [debounced, setDebounced] = useState(value);

  useEffect(() => {
    const timer = window.setTimeout(() => setDebounced(value), delayMs);
    return () => window.clearTimeout(timer);
  }, [value, delayMs]);

  return debounced;
}

function toJobListParams(
  filters: BoardFilters,
  sort: BoardSort,
  offset: number,
  limit: number,
): JobListParams {
  return {
    q: filters.q.trim() || undefined,
    type: filters.type || undefined,
    location: filters.location.trim() || undefined,
    recency: filters.recency || undefined,
    saved: filters.saved || undefined,
    sort: sort.field,
    order: sort.dir,
    limit,
    offset,
  };
}

export function useBoardListings(
  filters: BoardFilters,
  sort: BoardSort,
  offset: number,
  enabled: boolean,
) {
  const debouncedQ = useDebouncedValue(filters.q, 300);
  const debouncedLocation = useDebouncedValue(filters.location, 300);
  const params = toJobListParams(
    { ...filters, q: debouncedQ, location: debouncedLocation },
    sort,
    offset,
    BOARD_PAGE_SIZE,
  );
  const query = useJobs(params, enabled);
  const locationsQuery = useJobLocations(enabled);

  const listings = useMemo(
    () => (query.data?.jobs ?? []).map(jobToListing),
    [query.data],
  );

  const facets = useMemo<BoardFacets>(
    () => ({
      types: uniqueSorted([
        ...DEFAULT_TYPES,
        ...listings
          .map((listing) => listing.type)
          .filter((type) => type && type !== "—"),
        filters.type,
      ]),
      locations: uniqueSorted(locationsQuery.data ?? []),
    }),
    [listings, filters.type, locationsQuery.data],
  );

  return {
    listings,
    total: query.data?.total ?? 0,
    limit: query.data?.limit ?? BOARD_PAGE_SIZE,
    offset: query.data?.offset ?? offset,
    isPending: query.isPending,
    isFetching: query.isFetching,
    isError: query.isError,
    facets,
  };
}
