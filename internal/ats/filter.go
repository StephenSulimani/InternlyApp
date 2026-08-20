package ats

import (
	"regexp"
	"strings"
)

var (
	internTitleRE  = regexp.MustCompile(`(?i)\bintern(?:s|ship|ships)?\b|\bco-?ops?\b`)
	newGradTitleRE = regexp.MustCompile(`(?i)\bnew[\s-]*grads?(?:uate)?s?\b|\bearly[\s-]*career\b|\buniversity\s+(?:grad|hire|recruit)|campus\s+(?:hire|recruit)`)
)

// ClassifyEarlyCareer returns Internship or Full Time when the title looks
// like an intern / new-grad / early-career role. ok is false for everything else.
func ClassifyEarlyCareer(title string) (jobType string, ok bool) {
	title = strings.TrimSpace(title)
	if title == "" {
		return "", false
	}
	if internTitleRE.MatchString(title) {
		return "Internship", true
	}
	if newGradTitleRE.MatchString(title) {
		return "Full Time", true
	}
	return "", false
}
