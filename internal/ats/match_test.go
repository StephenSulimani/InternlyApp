package ats

import "testing"

func TestApplicationMatchKeys(t *testing.T) {
	keys := ApplicationMatchKeys("https://boards.greenhouse.io/stripe/jobs/12345/")
	want := map[string]bool{
		"https://boards.greenhouse.io/stripe/jobs/12345/":    true,
		"https://boards.greenhouse.io/stripe/jobs/12345":     true,
		"https://job-boards.greenhouse.io/stripe/jobs/12345": true,
	}
	for _, key := range keys {
		delete(want, key)
	}
	if len(want) != 0 {
		t.Fatalf("missing keys %v (got %v)", want, keys)
	}
}

func TestApplicationMatchRegex(t *testing.T) {
	re := ApplicationMatchRegex("https://boards.greenhouse.io/stripe/jobs/12345")
	if re == "" {
		t.Fatal("expected regex")
	}
	if ApplicationMatchRegex("https://jobs.lever.co/openai/abcd") != "" {
		t.Fatal("lever urls should not produce a greenhouse id regex")
	}
	if greenhouseJobID("https://boards.greenhouse.io/embed/job_app?for=figma&token=999") != "999" {
		t.Fatal("expected embed token id")
	}
}
