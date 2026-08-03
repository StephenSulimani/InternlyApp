import type { Job } from "../api/jobs";

export type Listing = {
  id: string;
  company: string;
  role: string;
  location: string;
  type: string;
  posted: string;
  isNew?: boolean;
  applicationLink?: string;
};

export function jobToListing(job: Job): Listing {
  return {
    id: job.id,
    company: job.company || "Unknown",
    role: job.role_title || "Open role",
    location: job.locations?.length ? job.locations.join(" · ") : "—",
    type: job.job_type || "—",
    posted: formatRelativePosted(job.first_seen),
    isNew: isWithinHours(job.first_seen, 48),
    applicationLink: job.application_link,
  };
}

function formatRelativePosted(iso: string | undefined): string {
  if (!iso) {
    return "—";
  }

  const then = new Date(iso).getTime();
  if (Number.isNaN(then)) {
    return "—";
  }

  const diffMs = Date.now() - then;
  const minutes = Math.floor(diffMs / 60_000);
  if (minutes < 1) {
    return "Just now";
  }
  if (minutes < 60) {
    return `${minutes}m ago`;
  }

  const hours = Math.floor(minutes / 60);
  if (hours < 24) {
    return `${hours}h ago`;
  }

  const days = Math.floor(hours / 24);
  if (days === 1) {
    return "Yesterday";
  }
  return `${days}d ago`;
}

function isWithinHours(iso: string | undefined, hours: number): boolean {
  if (!iso) {
    return false;
  }
  const then = new Date(iso).getTime();
  if (Number.isNaN(then)) {
    return false;
  }
  return Date.now() - then <= hours * 60 * 60 * 1000;
}
