# Internly Web

React frontend for Internly.

## Development

```bash
cd web
npm install
cp .env.example .env
npm run dev
```

Open [http://localhost:5173](http://localhost:5173).

## Build

```bash
npm run build
npm run preview
```

## Data fetching (TanStack Query)

API calls live under `src/api/`. React Query hooks live under `src/hooks/`.

| Path | Purpose |
|------|---------|
| `src/lib/queryClient.ts` | Shared `QueryClient` defaults |
| `src/api/client.ts` | Base `fetch` wrapper + `ApiError` |
| `src/api/queryKeys.ts` | Central query key factories |
| `src/api/jobs.ts` | Job endpoints (stubbed until API exists) |
| `src/hooks/useJobs.ts` | `useJobs` / `useJob` hooks |

Hooks are **disabled by default** (`VITE_ENABLE_API=false`). The homepage still uses mock data. When the Go API exposes job routes, set `VITE_ENABLE_API=true` and swap components to `useJobs()`.

TanStack Query Devtools appear in development (bottom-left).
