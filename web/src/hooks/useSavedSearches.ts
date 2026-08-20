import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
  createSavedSearch,
  deleteSavedSearch,
  fetchSavedSearches,
  updateSavedSearch,
  type SavedSearchInput,
} from "../api/savedSearches";
import { queryKeys } from "../api/queryKeys";

export function useSavedSearches(enabled: boolean) {
  return useQuery({
    queryKey: queryKeys.savedSearches.list(),
    queryFn: fetchSavedSearches,
    enabled,
    staleTime: 30_000,
  });
}

export function useSavedSearchMutations() {
  const queryClient = useQueryClient();

  const invalidate = () =>
    void queryClient.invalidateQueries({ queryKey: queryKeys.savedSearches.all });

  const create = useMutation({
    mutationFn: (input: SavedSearchInput) => createSavedSearch(input),
    onSuccess: invalidate,
  });

  const update = useMutation({
    mutationFn: ({ id, input }: { id: string; input: SavedSearchInput }) =>
      updateSavedSearch(id, input),
    onSuccess: invalidate,
  });

  const remove = useMutation({
    mutationFn: (id: string) => deleteSavedSearch(id),
    onSuccess: invalidate,
  });

  return { create, update, remove };
}
