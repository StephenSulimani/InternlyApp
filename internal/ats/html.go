package ats

import (
	"html"
	"regexp"
	"strings"
)

var (
	htmlScript      = regexp.MustCompile(`(?is)<script[^>]*>.*?</script>`)
	htmlStyle       = regexp.MustCompile(`(?is)<style[^>]*>.*?</style>`)
	htmlComments    = regexp.MustCompile(`(?s)<!--.*?-->`)
	htmlBreaks      = regexp.MustCompile(`(?i)<br\s*/?>|</(p|div|h[1-6]|li|tr|blockquote|ul|ol|table|section|article|header|footer|pre)>`)
	htmlTags        = regexp.MustCompile(`<[^>]+>`)
	blankLines      = regexp.MustCompile(`\n{3,}`)
)

// HTMLToText turns ATS HTML job content into readable plain text for storage and search.
func HTMLToText(raw string) string {
	s := strings.ReplaceAll(raw, "\u00a0", " ")
	for i := 0; i < 3; i++ {
		s = html.UnescapeString(s)
	}
	s = htmlScript.ReplaceAllString(s, "")
	s = htmlStyle.ReplaceAllString(s, "")
	s = htmlComments.ReplaceAllString(s, "")
	s = htmlBreaks.ReplaceAllString(s, "\n")
	for strings.Contains(s, "<") {
		prev := s
		s = htmlTags.ReplaceAllString(s, "")
		if s == prev {
			break
		}
	}
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
