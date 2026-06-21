import { QueryClientProvider } from "@tanstack/react-query";
import { ReactQueryDevtools } from "@tanstack/react-query-devtools";
import { BrowserRouter } from "react-router-dom";
import { queryClient } from "../lib/queryClient";

type Props = {
  children: React.ReactNode;
};

export default function AppProviders({ children }: Props) {
  return (
    <QueryClientProvider client={queryClient}>
      <BrowserRouter>{children}</BrowserRouter>
      {import.meta.env.DEV && (
        <ReactQueryDevtools buttonPosition="bottom-left" initialIsOpen={false} />
      )}
    </QueryClientProvider>
  );
}
