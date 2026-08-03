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
};

/** GET /jobs */
export async function fetchJobs(limit = 50): Promise<Job[]> {
  const res = await apiClient<ApiResponse<Job[]>>(`/jobs?limit=${limit}`);
  return res.data ?? [];
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

/** GET /jobs/:id (not implemented on API yet). */
export async function fetchJob(id: string): Promise<Job> {
  const res = await apiClient<ApiResponse<Job>>(`/jobs/${id}`);
  if (!res.data) {
    throw new Error("Job not found");
  }
  return res.data;
}
