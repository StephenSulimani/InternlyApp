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
  recency: RecencyFilter;
  saved: boolean;
};

export const EMPTY_BOARD_FILTERS: BoardFilters = {
  q: "",
  type: "",
  location: "",
  recency: "",
  saved: false,
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
  const recencyMs = filters.recency ? RECENCY_MS[filters.recency] : null;
  const now = Date.now();

  return listings.filter((listing) => {
    const queries = splitLocationTokens(filters.q).map((token) =>
      token.toLowerCase(),
    );
    if (queries.length > 0) {
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
      if (!queries.some((needle) => haystack.includes(needle))) {
        return false;
      }
    }

    if (filters.type && listing.type !== filters.type) {
      return false;
    }

    if (filters.location) {
      const needles = splitLocationTokens(filters.location).map((token) =>
        token.toLowerCase(),
      );
      if (needles.length > 0) {
        const haystack = [...listing.locations, listing.location]
          .filter(Boolean)
          .map((loc) => loc.toLowerCase());
        const matched = needles.some((needle) =>
          haystack.some((loc) => loc.includes(needle)),
        );
        if (!matched) {
          return false;
        }
      }
    }

    if (filters.saved && !listing.saved) {
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
    filters.recency !== "" ||
    filters.saved
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

/**
 * CSV location tokens. Values with commas are quoted so
 * `"New York, NY"` stays one location.
 */
export function splitLocationTokens(value: string): string[] {
  const tokens: string[] = [];
  let current = "";
  let inQuotes = false;

  for (let i = 0; i < value.length; i += 1) {
    const char = value[i];
    if (inQuotes) {
      if (char === '"') {
        if (value[i + 1] === '"') {
          current += '"';
          i += 1;
        } else {
          inQuotes = false;
        }
      } else {
        current += char;
      }
      continue;
    }

    if (char === '"') {
      inQuotes = true;
      continue;
    }

    if (char === ",") {
      const token = current.trim();
      if (token) {
        tokens.push(token);
      }
      current = "";
      continue;
    }

    current += char;
  }

  const token = current.trim();
  if (token) {
    tokens.push(token);
  }
  return tokens;
}

export function joinLocationTokens(tokens: string[]): string {
  return tokens
    .map((token) => token.trim())
    .filter(Boolean)
    .map(quoteLocationToken)
    .join(", ");
}

function quoteLocationToken(token: string): string {
  if (/[",\n\r]/.test(token)) {
    return `"${token.replaceAll('"', '""')}"`;
  }
  return token;
}

export function addLocationToken(selected: string[], token: string): string[] {
  const next = token.trim();
  if (!next) {
    return selected;
  }
  if (selected.some((item) => item.toLowerCase() === next.toLowerCase())) {
    return selected;
  }
  return [...selected, next];
}

export function removeLocationToken(selected: string[], token: string): string[] {
  return selected.filter((item) => item.toLowerCase() !== token.toLowerCase());
}

export function matchingLocationSuggestions(
  locations: string[],
  selected: string[],
  draft: string,
  limit = 8,
): string[] {
  const selectedSet = new Set(selected.map((token) => token.toLowerCase()));
  const needle = draft.trim().toLowerCase();

  return locations
    .filter((location) => {
      if (selectedSet.has(location.toLowerCase())) {
        return false;
      }
      if (!needle) {
        return true;
      }
      return location.toLowerCase().includes(needle);
    })
    .slice(0, limit);
}
