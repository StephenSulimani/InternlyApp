import type { Listing } from "./listings";

export type RecencyFilter = "" | "24h" | "3d" | "7d";

export type BoardSortField = "posted" | "company" | "role" | "location" | "type";
export type BoardSortDir = "asc" | "desc";

export type BoardSort = {
  field: BoardSortField;
  dir: BoardSortDir;
};

export const DEFAULT_BOARD_SORT: BoardSort = {
  field: "posted",
  dir: "desc",
};

/** Client-side board filters — aligned with GET /jobs query params. */
export type BoardFilters = {
  q: string;
  type: string;
  location: string;
  source: string;
  recency: RecencyFilter;
};

export const EMPTY_BOARD_FILTERS: BoardFilters = {
  q: "",
  type: "",
  location: "",
  source: "",
  recency: "",
};

const RECENCY_MS: Record<Exclude<RecencyFilter, "">, number> = {
  "24h": 24 * 60 * 60 * 1000,
  "3d": 3 * 24 * 60 * 60 * 1000,
  "7d": 7 * 24 * 60 * 60 * 1000,
};

export function filterListings(
  listings: Listing[],
  filters: BoardFilters,
): Listing[] {
  const q = filters.q.trim().toLowerCase();
  const recencyMs = filters.recency ? RECENCY_MS[filters.recency] : null;
  const now = Date.now();

  return listings.filter((listing) => {
    if (q) {
      const haystack = [
        listing.company,
        listing.role,
        listing.location,
        listing.type,
        listing.source,
        listing.description,
      ]
        .filter(Boolean)
        .join(" ")
        .toLowerCase();
      if (!haystack.includes(q)) {
        return false;
      }
    }

    if (filters.type && listing.type !== filters.type) {
      return false;
    }

    if (filters.location) {
      const inLocations = listing.locations.includes(filters.location);
      const inLabel = listing.location === filters.location;
      if (!inLocations && !inLabel) {
        return false;
      }
    }

    if (filters.source && listing.source !== filters.source) {
      return false;
    }

    if (recencyMs != null) {
      if (!listing.firstSeen) {
        return false;
      }
      const then = new Date(listing.firstSeen).getTime();
      if (Number.isNaN(then) || now - then > recencyMs) {
        return false;
      }
    }

    return true;
  });
}

export function hasActiveFilters(filters: BoardFilters): boolean {
  return (
    filters.q.trim() !== "" ||
    filters.type !== "" ||
    filters.location !== "" ||
    filters.source !== "" ||
    filters.recency !== ""
  );
}

export function uniqueSorted(values: string[]): string[] {
  return [...new Set(values.filter(Boolean))].sort((a, b) =>
    a.localeCompare(b),
  );
}

export function listingTypes(listings: Listing[]): string[] {
  return uniqueSorted(listings.map((listing) => listing.type));
}

export function listingLocations(listings: Listing[]): string[] {
  return uniqueSorted(listings.flatMap((listing) => listing.locations));
}

export function listingSources(listings: Listing[]): string[] {
  return uniqueSorted(
    listings.map((listing) => listing.source).filter((s): s is string => Boolean(s)),
  );
}
