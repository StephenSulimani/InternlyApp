import { apiClient } from "./client";

/** Matches the Go API response envelope once wired up. */
export type ApiResponse<T = unknown> = {
  success: boolean;
  message: string;
  data?: T;
};

/** Job listing shape — aligned with cmd/api/types.JobListing. */
export type Job = {
  id: string;
  company: string;
  role_title: string;
  locations: string[];
  job_type: string;
  application_link: string;
  first_seen?: string;
  source_name: string;
  description?: string;
  saved?: boolean;
};

/** Query params for GET /jobs. */
export type JobListParams = {
  /** CSV search terms (quote tokens that contain commas); a job matches if any term matches. */
  q?: string;
  type?: string;
  /** CSV locations (quote tokens that contain commas); a job matches if any token matches. */
  location?: string;
  source?: string;
  recency?: "24h" | "3d" | "7d" | "";
  saved?: boolean;
  sort?: "posted" | "company" | "role" | "location" | "type";
  order?: "asc" | "desc";
  limit?: number;
  offset?: number;
};

/** Paginated GET /jobs payload. */
export type JobsPage = {
  jobs: Job[];
  total: number;
  limit: number;
  offset: number;
};

/** GET /jobs */
export async function fetchJobs(params: JobListParams = {}): Promise<JobsPage> {
  const query = new URLSearchParams();
  if (params.q) {
    query.set("q", params.q);
  }
  if (params.type) {
    query.set("type", params.type);
  }
  if (params.location) {
    query.set("location", params.location);
  }
  if (params.source) {
    query.set("source", params.source);
  }
  if (params.recency) {
    query.set("recency", params.recency);
  }
  if (params.saved) {
    query.set("saved", "true");
  }
  if (params.sort) {
    query.set("sort", params.sort);
  }
  if (params.order) {
    query.set("order", params.order);
  }
  if (params.limit != null) {
    query.set("limit", String(params.limit));
  }
  if (params.offset != null) {
    query.set("offset", String(params.offset));
  }

  const qs = query.toString();
  const res = await apiClient<ApiResponse<JobsPage>>(`/jobs${qs ? `?${qs}` : ""}`);
  return (
    res.data ?? {
      jobs: [],
      total: 0,
      limit: params.limit ?? 50,
      offset: params.offset ?? 0,
    }
  );
}

/** Board aggregate stats from GET /jobs/stats. */
export type JobStats = {
  total_jobs: number;
  added_this_week: number;
  total_companies: number;
  last_updated?: string;
};

/** GET /jobs/stats */
export async function fetchJobStats(): Promise<JobStats> {
  const res = await apiClient<ApiResponse<JobStats>>("/jobs/stats");
  if (!res.data) {
    throw new Error("Job stats unavailable");
  }
  return res.data;
}

/** GET /board/preview — public teaser listings for the homepage. */
export async function fetchBoardPreview(): Promise<Job[]> {
  const res = await apiClient<ApiResponse<Job[]>>("/board/preview");
  return res.data ?? [];
}

/** GET /jobs/locations — distinct locations that have at least one job. */
export async function fetchJobLocations(): Promise<string[]> {
  const res = await apiClient<ApiResponse<string[]>>("/jobs/locations");
  return res.data ?? [];
}

/** PUT /jobs/:id/save */
export async function saveJob(id: string): Promise<boolean> {
  const res = await apiClient<ApiResponse<{ saved: boolean }>>(
    `/jobs/${id}/save`,
    { method: "PUT" },
  );
  return res.data?.saved ?? true;
}

/** DELETE /jobs/:id/save */
export async function unsaveJob(id: string): Promise<boolean> {
  const res = await apiClient<ApiResponse<{ saved: boolean }>>(
    `/jobs/${id}/save`,
    { method: "DELETE" },
  );
  return res.data?.saved ?? false;
}

/** GET /jobs/:id (not implemented on API yet). */
export async function fetchJob(id: string): Promise<Job> {
  const res = await apiClient<ApiResponse<Job>>(`/jobs/${id}`);
  if (!res.data) {
    throw new Error("Job not found");
  }
  return res.data;
}
