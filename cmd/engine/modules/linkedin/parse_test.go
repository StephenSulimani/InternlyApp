package linkedin

import (
	"encoding/json"
	"testing"
)

func TestCleanJobURL(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"strips query params", "https://www.linkedin.com/jobs/view/123?refId=abc", "https://www.linkedin.com/jobs/view/123"},
		{"unchanged without query", "https://www.linkedin.com/jobs/view/123", "https://www.linkedin.com/jobs/view/123"},
		{"empty string", "", ""},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := CleanJobURL(tc.in); got != tc.want {
				t.Fatalf("CleanJobURL(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestBuildJobMetadata(t *testing.T) {
	raw, err := BuildJobMetadata("Entry level", "Build trading systems", "https://cdn.example/logo.png")
	if err != nil {
		t.Fatal(err)
	}

	var metadata map[string]string
	if err := json.Unmarshal(raw, &metadata); err != nil {
		t.Fatal(err)
	}

	if metadata["seniority"] != "Entry level" {
		t.Fatalf("seniority = %q", metadata["seniority"])
	}
	if metadata["description"] != "Build trading systems" {
		t.Fatalf("description = %q", metadata["description"])
	}
	if metadata["company_image_url"] != "https://cdn.example/logo.png" {
		t.Fatalf("company_image_url = %q", metadata["company_image_url"])
	}
}
