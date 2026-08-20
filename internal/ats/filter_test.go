package ats

import "testing"

func TestClassifyEarlyCareer(t *testing.T) {
	cases := []struct {
		title   string
		jobType string
		ok      bool
	}{
		{"Software Engineer Intern", "Internship", true},
		{"Intern, Data Science", "Internship", true},
		{"Summer Internship 2026", "Internship", true},
		{"Software Co-op", "Internship", true},
		{"New Grad Software Engineer", "Full Time", true},
		{"Early Career SWE", "Full Time", true},
		{"University Graduate Engineer", "Full Time", true},
		{"Campus Hire - Software", "Full Time", true},
		{"Staff Software Engineer", "", false},
		{"International Expansion Lead", "", false},
		{"Internal Tools Engineer", "", false},
		{"", "", false},
	}
	for _, tc := range cases {
		got, ok := ClassifyEarlyCareer(tc.title)
		if ok != tc.ok || got != tc.jobType {
			t.Fatalf("ClassifyEarlyCareer(%q) = (%q, %v), want (%q, %v)", tc.title, got, ok, tc.jobType, tc.ok)
		}
	}
}
