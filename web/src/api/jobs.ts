import { apiClient } from "./client";

/** Matches the Go API response envelope once wired up. */
export type ApiResponse<T = unknown> = {
  success: boolean;
  message: string;
  data?: T;
};

/** Job listing shape — align with internal/db.Job when integrating. */
export type Job = {
  id: string;
  company: string;
  roleTitle: string;
  location: string;
  jobType: string;
  applicationLink: string;
  postedAt?: string;
};

/** GET /jobs (not implemented on API yet). */
export async function fetchJobs(): Promise<Job[]> {
  const res = await apiClient<ApiResponse<Job[]>>("/jobs");
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
