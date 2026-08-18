import {
  DEFAULT_BOARD_SORT,
  EMPTY_BOARD_FILTERS,
  type BoardFilters,
  type BoardSort,
  type BoardSortDir,
  type BoardSortField,
  type RecencyFilter,
} from "./boardFilters";

const STORAGE_PREFIX = "internly_board_search.v1";

const RECENCY: ReadonlySet<string> = new Set(["", "24h", "3d", "7d"]);
const SORT_FIELDS: ReadonlySet<string> = new Set([
  "posted",
  "company",
  "role",
  "location",
  "type",
]);
const SORT_DIRS: ReadonlySet<string> = new Set(["asc", "desc"]);

export type BoardSearchState = {
  filters: BoardFilters;
  sort: BoardSort;
};

export const EMPTY_BOARD_SEARCH: BoardSearchState = {
  filters: EMPTY_BOARD_FILTERS,
  sort: DEFAULT_BOARD_SORT,
};

export function boardSearchStorageKey(userId: string): string {
  return `${STORAGE_PREFIX}:${userId}`;
}

export function loadBoardSearch(userId: string | undefined): BoardSearchState {
  if (!userId || typeof localStorage === "undefined") {
    return EMPTY_BOARD_SEARCH;
  }

  const raw = localStorage.getItem(boardSearchStorageKey(userId));
  if (!raw) {
    return EMPTY_BOARD_SEARCH;
  }

  try {
    const parsed = JSON.parse(raw) as unknown;
    return parseBoardSearch(parsed);
  } catch {
    return EMPTY_BOARD_SEARCH;
  }
}

export function saveBoardSearch(
  userId: string | undefined,
  state: BoardSearchState,
): void {
  if (!userId || typeof localStorage === "undefined") {
    return;
  }

  try {
    localStorage.setItem(boardSearchStorageKey(userId), JSON.stringify(state));
  } catch {
    // Quota or private-mode restrictions — search still works for this visit.
  }
}

function parseBoardSearch(value: unknown): BoardSearchState {
  if (!value || typeof value !== "object") {
    return EMPTY_BOARD_SEARCH;
  }

  const record = value as Record<string, unknown>;
  return {
    filters: parseBoardFilters(record.filters),
    sort: parseBoardSort(record.sort),
  };
}

function parseBoardFilters(value: unknown): BoardFilters {
  if (!value || typeof value !== "object") {
    return EMPTY_BOARD_FILTERS;
  }

  const record = value as Record<string, unknown>;
  const recency = record.recency;
  return {
    q: asString(record.q),
    type: asString(record.type),
    location: asString(record.location),
    source: asString(record.source),
    recency: RECENCY.has(String(recency ?? ""))
      ? (recency as RecencyFilter)
      : "",
    saved: record.saved === true,
  };
}

function parseBoardSort(value: unknown): BoardSort {
  if (!value || typeof value !== "object") {
    return DEFAULT_BOARD_SORT;
  }

  const record = value as Record<string, unknown>;
  const field = record.field;
  const dir = record.dir;
  if (
    typeof field !== "string" ||
    !SORT_FIELDS.has(field) ||
    typeof dir !== "string" ||
    !SORT_DIRS.has(dir)
  ) {
    return DEFAULT_BOARD_SORT;
  }

  return { field: field as BoardSortField, dir: dir as BoardSortDir };
}

function asString(value: unknown): string {
  return typeof value === "string" ? value : "";
}
