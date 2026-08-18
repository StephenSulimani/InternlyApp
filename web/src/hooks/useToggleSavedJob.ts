import { useMutation, useQueryClient } from "@tanstack/react-query";
import { saveJob, unsaveJob } from "../api/jobs";
import { queryKeys } from "../api/queryKeys";

export function useToggleSavedJob() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async ({ id, saved }: { id: string; saved: boolean }) => {
      if (saved) {
        await unsaveJob(id);
        return false;
      }
      await saveJob(id);
      return true;
    },
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: queryKeys.jobs.all });
    },
  });
}
