package linkedin

import (
	"strings"
	"testing"
)

func TestLinkedInSearchParams_BuildURL(t *testing.T) {
	t.Run("keyword and location only", func(t *testing.T) {
		params := LinkedInSearchParams{
			Keyword:  "Software Engineer",
			Location: "New York",
		}

		got := params.BuildURL()
		want := "https://www.linkedin.com/jobs/search?keywords=Software Engineer&location=New York"
		if got != want {
			t.Fatalf("url = %q, want %q", got, want)
		}
	})

	t.Run("includes optional filters", func(t *testing.T) {
		params := LinkedInSearchParams{
			Keyword:         "Quantitative Developer",
			Location:        "New York",
			JobType:         []JobType{FullTime, Contract},
			ExperienceLevel: []ExperienceLevel{EntryLevel, MidSenior},
			Salary:          Over100k,
			WorkType:        []WorkType{OnSite, Hybrid},
			DatePosted:      LastWeek,
		}

		got := params.BuildURL()
		want := "https://www.linkedin.com/jobs/search?keywords=Quantitative Developer&location=New York" +
			"&f_JT=F,C&f_E=2,4&f_SB2=4&f_WT=1,3&f_TPR=r604800"
		if got != want {
			t.Fatalf("url = %q, want %q", got, want)
		}
	})

	t.Run("omits empty optional filters", func(t *testing.T) {
		params := LinkedInSearchParams{
			Keyword:  "Intern",
			Location: "Remote",
		}

		got := params.BuildURL()
		if strings.Contains(got, "f_JT=") || strings.Contains(got, "f_E=") || strings.Contains(got, "f_SB2=") ||
			strings.Contains(got, "f_WT=") || strings.Contains(got, "f_TPR=") {
			t.Fatalf("expected no filter params, got %q", got)
		}
	})
}
