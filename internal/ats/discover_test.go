package ats

import "testing"

func TestDiscover(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want Board
		ok   bool
	}{
		{
			name: "greenhouse job",
			in:   "https://boards.greenhouse.io/stripe/jobs/12345",
			want: Board{Name: "greenhouse", URL: "https://boards.greenhouse.io/stripe"},
			ok:   true,
		},
		{
			name: "greenhouse job-boards host",
			in:   "https://job-boards.greenhouse.io/andurilindustries/jobs/4603607007",
			want: Board{Name: "greenhouse", URL: "https://job-boards.greenhouse.io/andurilindustries"},
			ok:   true,
		},
		{
			name: "greenhouse embed",
			in:   "https://boards.greenhouse.io/embed/job_app?for=figma&token=123",
			want: Board{Name: "greenhouse", URL: "https://boards.greenhouse.io/figma"},
			ok:   true,
		},
		{
			name: "lever job",
			in:   "https://jobs.lever.co/openai/abcd-efgh",
			want: Board{Name: "lever", URL: "https://jobs.lever.co/openai"},
			ok:   true,
		},
		{
			name: "ashby job",
			in:   "https://jobs.ashbyhq.com/anthropic/role-id",
			want: Board{Name: "ashby", URL: "https://jobs.ashbyhq.com/anthropic"},
			ok:   true,
		},
		{
			name: "workday job",
			in:   "https://nvidia.wd5.myworkdayjobs.com/en-US/NVIDIAExternalCareerSite/job/Santa-Clara/Software-Engineer_JR123",
			want: Board{Name: "workday", URL: "https://nvidia.wd5.myworkdayjobs.com/en-US/NVIDIAExternalCareerSite"},
			ok:   true,
		},
		{
			name: "smartrecruiters",
			in:   "https://jobs.smartrecruiters.com/Spotify/743999999",
			want: Board{Name: "smartrecruiters", URL: "https://jobs.smartrecruiters.com/Spotify"},
			ok:   true,
		},
		{
			name: "workable",
			in:   "https://apply.workable.com/acme/j/ABC123/",
			want: Board{Name: "workable", URL: "https://apply.workable.com/acme"},
			ok:   true,
		},
		{
			name: "icims",
			in:   "https://careers-acme.icims.com/jobs/1234/job",
			want: Board{Name: "icims", URL: "https://careers-acme.icims.com/jobs"},
			ok:   true,
		},
		{
			name: "not an ats",
			in:   "https://www.linkedin.com/jobs/view/123",
			ok:   false,
		},
		{
			name: "empty",
			in:   "",
			ok:   false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := Discover(tc.in)
			if ok != tc.ok {
				t.Fatalf("ok = %v, want %v (got %+v)", ok, tc.ok, got)
			}
			if !tc.ok {
				return
			}
			if got != tc.want {
				t.Fatalf("board = %+v, want %+v", got, tc.want)
			}
		})
	}
}
