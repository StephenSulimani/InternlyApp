import { useMutation } from "@tanstack/react-query";
import { login, saveToken, type LoginInput } from "../api/auth";

export function useLogin() {
  return useMutation({
    mutationFn: (input: LoginInput) => login(input),
    onSuccess: (data) => {
      saveToken(data.token);
    },
  });
}
