import type { Job } from "../api/jobs";
import { jobToListing } from "../lib/listings";

function hoursAgo(hours: number): string {
  return new Date(Date.now() - hours * 60 * 60 * 1000).toISOString();
}

/** Placeholder catalog until GET /jobs search/filter is wired. */
export const mockJobs: Job[] = [
  {
    id: "job-stripe-swe-intern",
    company: "Stripe",
    role_title: "Software Engineer Intern",
    locations: ["San Francisco, CA", "Remote"],
    job_type: "Internship",
    application_link: "https://boards.greenhouse.io/stripe/jobs/12345",
    first_seen: hoursAgo(2),
    source_name: "Greenhouse",
    description:
      "Build payments infrastructure used by millions of businesses. You'll ship production code with a mentor, rotate through a team, and leave with a project you can point to.",
  },
  {
    id: "job-jane-street-qr",
    company: "Jane Street",
    role_title: "Quantitative Researcher",
    locations: ["New York, NY"],
    job_type: "Full Time",
    application_link:
      "https://www.janestreet.com/join-jane-street/position/123",
    first_seen: hoursAgo(5),
    source_name: "Simplify",
    description:
      "Work on research problems that sit next to live trading. Strong probability, programming, and a taste for puzzles matter more than a specific major.",
  },
  {
    id: "job-two-sigma-eng-intern",
    company: "Two Sigma",
    role_title: "Engineering Intern",
    locations: ["New York, NY"],
    job_type: "Internship",
    application_link: "https://careers.twosigma.com/careers/JobDetail/456",
    first_seen: hoursAgo(20),
    source_name: "Workday",
    description:
      "Join a small pod building data platforms and research tools. Interns own well-scoped projects and present to the broader engineering org at the end of summer.",
  },
  {
    id: "job-notion-pd",
    company: "Notion",
    role_title: "Product Designer",
    locations: ["San Francisco, CA", "New York, NY", "Remote"],
    job_type: "Full Time",
    application_link: "https://jobs.ashbyhq.com/notion/abcd",
    first_seen: hoursAgo(26),
    source_name: "Ashby",
    description:
      "Design the surfaces millions of teams live in every day. You'll partner with PMs and engineers from research through high-fidelity UI.",
  },
  {
    id: "job-anthropic-research-intern",
    company: "Anthropic",
    role_title: "Research Intern",
    locations: ["San Francisco, CA"],
    job_type: "Internship",
    application_link: "https://job-boards.greenhouse.io/anthropic/jobs/789",
    first_seen: hoursAgo(40),
    source_name: "Greenhouse",
    description:
      "Contribute to empirical research on large language models — evals, interpretability, or alignment, depending on the team match.",
  },
  {
    id: "job-citadel-swe",
    company: "Citadel",
    role_title: "Software Engineer",
    locations: ["Chicago, IL", "New York, NY"],
    job_type: "Full Time",
    application_link:
      "https://www.citadel.com/careers/details/software-engineer/",
    first_seen: hoursAgo(72),
    source_name: "Simplify",
    description:
      "Build low-latency systems that sit close to markets. New grads join a class, then embed with a trading or platform team.",
  },
  {
    id: "job-openai-applied-intern",
    company: "OpenAI",
    role_title: "Applied AI Intern",
    locations: ["San Francisco, CA"],
    job_type: "Internship",
    application_link: "https://jobs.lever.co/openai/applied-intern",
    first_seen: hoursAgo(8),
    source_name: "Lever",
    description:
      "Prototype product features on top of frontier models. Comfort with Python, evals, and messy real-world data is a plus.",
  },
  {
    id: "job-figma-swe-intern",
    company: "Figma",
    role_title: "Software Engineering Intern",
    locations: ["San Francisco, CA", "New York, NY"],
    job_type: "Internship",
    application_link: "https://boards.greenhouse.io/figma/jobs/555",
    first_seen: hoursAgo(14),
    source_name: "Greenhouse",
    description:
      "Ship features on a multiplayer design tool. Interns work across the stack — canvas performance, editor UX, or platform APIs.",
  },
  {
    id: "job-databricks-ng",
    company: "Databricks",
    role_title: "Software Engineer, New Grad",
    locations: ["San Francisco, CA", "Seattle, WA", "Remote"],
    job_type: "Full Time",
    application_link:
      "https://www.databricks.com/company/careers/open-positions/123",
    first_seen: hoursAgo(30),
    source_name: "Greenhouse",
    description:
      "Join a team working on Spark, Unity Catalog, or the lakehouse control plane. Strong systems and distributed-data instincts help.",
  },
  {
    id: "job-airbnb-intern",
    company: "Airbnb",
    role_title: "Software Engineer Intern",
    locations: ["San Francisco, CA"],
    job_type: "Internship",
    application_link: "https://careers.airbnb.com/positions/swe-intern",
    first_seen: hoursAgo(50),
    source_name: "Ashby",
    description:
      "Work on guest or host product surfaces. You'll pair with a mentor, ship to production, and present at intern expo.",
  },
  {
    id: "job-spotify-data-intern",
    company: "Spotify",
    role_title: "Data Science Intern",
    locations: ["New York, NY", "Remote"],
    job_type: "Internship",
    application_link: "https://jobs.lever.co/spotify/data-intern",
    first_seen: hoursAgo(18),
    source_name: "Lever",
    description:
      "Analyze listener behavior and help teams decide what to build next. SQL, Python, and a point of view on experimentation are useful.",
  },
  {
    id: "job-nvidia-ng",
    company: "NVIDIA",
    role_title: "Systems Software Engineer",
    locations: ["Santa Clara, CA", "Austin, TX"],
    job_type: "Full Time",
    application_link:
      "https://nvidia.wd5.myworkdayjobs.com/NVIDIAExternalCareerSite",
    first_seen: hoursAgo(96),
    source_name: "Workday",
    description:
      "Work on CUDA, drivers, or compiler tooling. New grads who enjoy performance, C/C++, and low-level debugging thrive here.",
  },
  {
    id: "job-shopify-intern",
    company: "Shopify",
    role_title: "Backend Engineer Intern",
    locations: ["Remote"],
    job_type: "Internship",
    application_link: "https://www.shopify.com/careers/intern-backend",
    first_seen: hoursAgo(6),
    source_name: "Ashby",
    description:
      "Help merchants run their businesses on a remote-first team. Ruby, Rails, and distributed systems show up often — curiosity matters more.",
  },
  {
    id: "job-bloomberg-swe",
    company: "Bloomberg",
    role_title: "Software Engineer",
    locations: ["New York, NY", "London"],
    job_type: "Full Time",
    application_link: "https://careers.bloomberg.com/job/software-engineer",
    first_seen: hoursAgo(110),
    source_name: "Simplify",
    description:
      "Build the Terminal and the data products that sit underneath it. Early-career engineers rotate, then settle onto a product squad.",
  },
  {
    id: "job-cloudflare-intern",
    company: "Cloudflare",
    role_title: "Systems Intern",
    locations: ["Austin, TX", "Remote"],
    job_type: "Internship",
    application_link: "https://boards.greenhouse.io/cloudflare/jobs/888",
    first_seen: hoursAgo(11),
    source_name: "Greenhouse",
    description:
      "Work close to the edge — Workers, DNS, or network control planes. Comfort with Go, Rust, or C is helpful but not required.",
  },
  {
    id: "job-roblox-ng",
    company: "Roblox",
    role_title: "Engineer, New Grad",
    locations: ["San Mateo, CA"],
    job_type: "Full Time",
    application_link: "https://careers.roblox.com/jobs/new-grad-engineer",
    first_seen: hoursAgo(60),
    source_name: "Greenhouse",
    description:
      "Build the engine, economy, or creator tools used by millions of experiences. Game or graphics background is a plus, not a gate.",
  },
  {
    id: "job-ramp-intern",
    company: "Ramp",
    role_title: "Software Engineer Intern",
    locations: ["New York, NY", "Remote"],
    job_type: "Internship",
    application_link: "https://jobs.ashbyhq.com/ramp/intern",
    first_seen: hoursAgo(3),
    source_name: "Ashby",
    description:
      "Ship spend-management features at a fast-moving fintech. Interns take real tickets, sit in design reviews, and demo to the company.",
  },
  {
    id: "job-duolingo-intern",
    company: "Duolingo",
    role_title: "Machine Learning Intern",
    locations: ["Pittsburgh, PA", "Remote"],
    job_type: "Internship",
    application_link: "https://careers.duolingo.com/jobs/ml-intern",
    first_seen: hoursAgo(44),
    source_name: "Lever",
    description:
      "Improve lesson ranking, speech, or personalization. Expect a mix of modeling, data work, and product sense.",
  },
];

export const mockListings = mockJobs.map(jobToListing);
