package linkedin

import (
	"encoding/json"
	"regexp"
)

var jobURLQueryPattern = regexp.MustCompile(`\?.*`)

func CleanJobURL(raw string) string {
	return jobURLQueryPattern.ReplaceAllString(raw, "")
}

func BuildJobMetadata(seniority, description, companyImageURL string) ([]byte, error) {
	return json.Marshal(map[string]string{
		"seniority":         seniority,
		"description":       description,
		"company_image_url": companyImageURL,
	})
}
