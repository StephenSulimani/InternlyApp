import { useEffect, useRef, useState } from "react";
import {
  type BoardFilters,
  type BoardSort,
} from "../lib/boardFilters";
import {
  EMPTY_BOARD_SEARCH,
  loadBoardSearch,
  saveBoardSearch,
} from "../lib/boardSearchStorage";

export function usePersistedBoardSearch(userId: string | undefined) {
  const userIdRef = useRef(userId);
  userIdRef.current = userId;

  const [filters, setFilters] = useState<BoardFilters>(
    () => loadBoardSearch(userId).filters,
  );
  const [sort, setSort] = useState<BoardSort>(
    () => loadBoardSearch(userId).sort,
  );
  const [activeSavedSearchId, setActiveSavedSearchId] = useState<string | null>(
    () => loadBoardSearch(userId).activeSavedSearchId ?? null,
  );

  useEffect(() => {
    const stored = userId ? loadBoardSearch(userId) : EMPTY_BOARD_SEARCH;
    setFilters(stored.filters);
    setSort(stored.sort);
    setActiveSavedSearchId(stored.activeSavedSearchId ?? null);
  }, [userId]);

  useEffect(() => {
    const id = userIdRef.current;
    if (!id) {
      return;
    }
    saveBoardSearch(id, { filters, sort, activeSavedSearchId });
  }, [filters, sort, activeSavedSearchId]);

  return {
    filters,
    setFilters,
    sort,
    setSort,
    activeSavedSearchId,
    setActiveSavedSearchId,
  };
}
