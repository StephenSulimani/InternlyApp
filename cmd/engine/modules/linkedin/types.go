package linkedin

import (
	"fmt"
	"strings"
)

type LinkedIn struct {
	ChromiumPath string
	SearchParams LinkedInSearchParams
}

type JobType string // f_JT

const (
	FullTime  JobType = "F"
	PartTime  JobType = "P"
	Contract  JobType = "C"
	Temporary JobType = "T"
	Volunteer JobType = "V"
)

type ExperienceLevel string // f_E

const (
	Internship ExperienceLevel = "1"
	EntryLevel ExperienceLevel = "2"
	Associate  ExperienceLevel = "3"
	MidSenior  ExperienceLevel = "4"
	Director   ExperienceLevel = "5"
)

type Salary string // f_SB2

const (
	Over40k  Salary = "1"
	Over60k  Salary = "2"
	Over80k  Salary = "3"
	Over100k Salary = "4"
	Over120k Salary = "5"
)

type DatePosted string // f_TPR

const (
	LastDay   DatePosted = "r86400"
	LastWeek  DatePosted = "r604800"
	LastMonth DatePosted = "r2592000"
	Anytime   DatePosted = ""
)

type WorkType string // f_WT

const (
	OnSite WorkType = "1"
	Remote WorkType = "2"
	Hybrid WorkType = "3"
)

type LinkedInSearchParams struct {
	Keyword         string
	Location        string
	JobType         []JobType
	ExperienceLevel []ExperienceLevel
	Salary          Salary
	WorkType        []WorkType
	DatePosted      DatePosted
}

func mapSlice[T ~string](input []T) []string {
	result := make([]string, len(input))
	for i, v := range input {
		result[i] = string(v)
	}
	return result
}

func (p *LinkedInSearchParams) BuildURL() string {
	baseUrl := fmt.Sprintf("https://www.linkedin.com/jobs/search?keywords=%s&location=%s", p.Keyword, p.Location)

	if len(p.JobType) > 0 {
		baseUrl += fmt.Sprintf("&f_JT=%s", strings.Join(mapSlice(p.JobType), ","))
	}

	if len(p.ExperienceLevel) > 0 {
		baseUrl += fmt.Sprintf("&f_E=%s", strings.Join(mapSlice(p.ExperienceLevel), ","))
	}

	if p.Salary != "" {
		baseUrl += fmt.Sprintf("&f_SB2=%s", p.Salary)
	}

	if len(p.WorkType) > 0 {
		baseUrl += fmt.Sprintf("&f_WT=%s", strings.Join(mapSlice(p.WorkType), ","))
	}

	if p.DatePosted != "" {
		baseUrl += fmt.Sprintf("&f_TPR=%s", p.DatePosted)
	}

	return baseUrl
}
