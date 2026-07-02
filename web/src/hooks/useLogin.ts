import { useMutation } from "@tanstack/react-query";
import { login, type LoginInput } from "../api/auth";
import { useAuth } from "../providers/AuthProvider";

export function useLogin() {
  const { setSession } = useAuth();

  return useMutation({
    mutationFn: (input: LoginInput) => login(input),
    onSuccess: (data) => {
      setSession(data.token, data.user);
    },
  });
}
