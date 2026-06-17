package linkedin

import (
	"fmt"
	"math/rand"
	"path/filepath"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/playwright-community/playwright-go"
	"github.com/stephensulimani/internlyapp/internal/db"
	"go.uber.org/zap"
)

func (l *LinkedIn) Scrape(log *zap.SugaredLogger) ([]db.Job, error) {
	jobs := []db.Job{}

	pw, err := playwright.Run()
	if err != nil {
		return nil, fmt.Errorf("start playwright: %w", err)
	}
	defer pw.Stop()

	absPath, err := filepath.Abs(l.ChromiumPath)
	if err != nil {
		return nil, fmt.Errorf("resolve chromium path: %w", err)
	}

	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		ExecutablePath: playwright.String(absPath),
		Headless:       playwright.Bool(l.Headless),
		Args: []string{
			"--disable-blink-features=AutomationControlled",
		},
	})
	if err != nil {
		return nil, fmt.Errorf("launch chromium at %s: %w", absPath, err)
	}
	defer browser.Close()

	page, err := browser.NewPage()
	if err != nil {
		return nil, fmt.Errorf("create page: %w", err)
	}

	url := l.SearchParams.BuildURL()

	if _, err = page.Goto(url); err != nil {
		return nil, fmt.Errorf("goto %s: %w", url, err)
	}

	element := page.Locator("#base-contextual-sign-in-modal > div > section > button > icon > svg")
	if _, err := element.IsVisible(); err == nil {
		element.Click()
	}

	listSelector := "#main-content > section.two-pane-serp-page__results-list > ul > li"
	entries := page.Locator(listSelector)

	count, err := entries.Count()
	if err != nil {
		return nil, fmt.Errorf("count job entries: %w", err)
	}
	if count == 0 {
		return nil, fmt.Errorf("no job listings found for %s", url)
	}

	for i := range count {
		if l.MaxJobs > 0 && i >= l.MaxJobs {
			break
		}

		sleepTime := rand.Intn(5)
		time.Sleep(time.Duration(sleepTime) * time.Second)

		entry := entries.Nth(i)
		entry.Click()

		title, _ := entry.Locator("h3").InnerText()
		company, _ := entry.Locator("h4").InnerText()
		location, _ := entry.Locator(".job-search-card__location").InnerText()
		companyImageURL, _ := entry.Locator(".artdeco-entity-image").GetAttribute("src")
		description, _ := page.Locator(".show-more-less-html__markup").InnerText()
		jobURL, _ := entry.Locator(".base-card__full-link").GetAttribute("href")
		jobURL = CleanJobURL(jobURL)

		jobFlags, _ := page.Locator(".description__job-criteria-item").All()
		jobType := ""
		seniority := ""

		for _, flag := range jobFlags {
			flagType, _ := flag.Locator("h3").InnerText()
			if strings.Contains(flagType, "Employment type") {
				jobType, _ = flag.Locator("span").InnerText()
				jobType = strings.TrimSpace(jobType)
			}
			if strings.Contains(flagType, "Seniority level") {
				seniority, _ = flag.Locator("span").InnerText()
				seniority = strings.TrimSpace(seniority)
			}
		}

		metadata, err := BuildJobMetadata(seniority, description, companyImageURL)
		if err != nil {
			log.Errorf("could not marshal metadata: %v", err)
		}

		jobs = append(jobs, db.Job{
			SourceUrl:       url,
			SourceName:      "LinkedIn",
			FirstSeen:       pgtype.Timestamptz{Time: time.Now(), Valid: true},
			ApplicationLink: jobURL,
			Company:         &company,
			RoleTitle:       &title,
			Locations:       []string{location},
			JobType:         &jobType,
			Metadata:        metadata,
		})
	}

	return jobs, nil
}
