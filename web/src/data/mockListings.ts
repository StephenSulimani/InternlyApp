export type Listing = {
  id: string;
  company: string;
  role: string;
  location: string;
  type: string;
  posted: string;
  isNew?: boolean;
};

export const mockListings: Listing[] = [
  {
    id: "1",
    company: "Stripe",
    role: "Software Engineer Intern",
    location: "San Francisco, CA",
    type: "Internship",
    posted: "2h ago",
    isNew: true,
  },
  {
    id: "2",
    company: "Jane Street",
    role: "Quantitative Researcher",
    location: "New York, NY",
    type: "New Grad",
    posted: "5h ago",
  },
  {
    id: "3",
    company: "Two Sigma",
    role: "Engineering Intern",
    location: "New York, NY",
    type: "Internship",
    posted: "Yesterday",
  },
  {
    id: "4",
    company: "Notion",
    role: "Product Designer",
    location: "Remote",
    type: "Full Time",
    posted: "Yesterday",
  },
  {
    id: "5",
    company: "Anthropic",
    role: "Research Intern",
    location: "San Francisco, CA",
    type: "Internship",
    posted: "2d ago",
  },
  {
    id: "6",
    company: "Citadel",
    role: "Software Engineer",
    location: "Chicago, IL",
    type: "New Grad",
    posted: "3d ago",
  },
];
