import type { SavedSearch, SavedSearchInput } from "../api/savedSearches";
import {
  DEFAULT_BOARD_SORT,
  EMPTY_BOARD_FILTERS,
  type BoardFilters,
  type BoardSort,
  type RecencyFilter,
} from "./boardFilters";

export function savedSearchToBoardState(search: SavedSearch): {
  filters: BoardFilters;
  sort: BoardSort;
} {
  return {
    filters: {
      q: search.q ?? "",
      type: search.type ?? "",
      location: search.location ?? "",
      recency: (search.recency ?? "") as RecencyFilter,
      saved: Boolean(search.saved),
    },
    sort: {
      field: search.sort ?? DEFAULT_BOARD_SORT.field,
      dir: search.order ?? DEFAULT_BOARD_SORT.dir,
    },
  };
}

export function boardStateToSavedSearchInput(
  name: string,
  filters: BoardFilters,
  sort: BoardSort,
): SavedSearchInput {
  return {
    name,
    q: filters.q,
    type: filters.type,
    location: filters.location,
    recency: filters.recency,
    saved: filters.saved,
    sort: sort.field,
    order: sort.dir,
  };
}

export function boardStateMatchesSavedSearch(
  filters: BoardFilters,
  sort: BoardSort,
  search: SavedSearch,
): boolean {
  const expected = savedSearchToBoardState(search);
  return (
    filters.q === expected.filters.q &&
    filters.type === expected.filters.type &&
    filters.location === expected.filters.location &&
    filters.recency === expected.filters.recency &&
    filters.saved === expected.filters.saved &&
    sort.field === expected.sort.field &&
    sort.dir === expected.sort.dir
  );
}

export function isEmptyBoardState(
  filters: BoardFilters,
  sort: BoardSort,
): boolean {
  return (
    filters.q === EMPTY_BOARD_FILTERS.q &&
    filters.type === EMPTY_BOARD_FILTERS.type &&
    filters.location === EMPTY_BOARD_FILTERS.location &&
    filters.recency === EMPTY_BOARD_FILTERS.recency &&
    filters.saved === EMPTY_BOARD_FILTERS.saved &&
    sort.field === DEFAULT_BOARD_SORT.field &&
    sort.dir === DEFAULT_BOARD_SORT.dir
  );
}
