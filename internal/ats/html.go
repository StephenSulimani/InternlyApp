package ats

import (
	"html"
	"regexp"
	"strings"
)

var (
	htmlBreaks = regexp.MustCompile(`(?i)<br\s*/?>|</(p|div|h[1-6]|li|tr|blockquote)>`)
	htmlTags   = regexp.MustCompile(`<[^>]+>`)
	blankLines = regexp.MustCompile(`\n{3,}`)
)

// HTMLToText turns ATS HTML job content into readable plain text for storage and search.
func HTMLToText(raw string) string {
	s := strings.ReplaceAll(raw, "\u00a0", " ")
	s = htmlBreaks.ReplaceAllString(s, "\n")
	s = htmlTags.ReplaceAllString(s, "")
	s = html.UnescapeString(s)
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")

	lines := strings.Split(s, "\n")
	for i, line := range lines {
		lines[i] = strings.TrimSpace(line)
	}
	s = strings.Join(lines, "\n")
	s = blankLines.ReplaceAllString(s, "\n\n")
	return strings.TrimSpace(s)
}
