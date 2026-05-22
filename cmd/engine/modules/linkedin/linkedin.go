package linkedin

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/playwright-community/playwright-go"
	"github.com/stephensulimani/internlyapp/internal/db"
	"go.uber.org/zap"
)

func (l *LinkedIn) Scrape(log *zap.SugaredLogger) []db.Job {
	jobs := []db.Job{}
	pw, err := playwright.Run()
	if err != nil {
		log.Fatalf("could not start playwright: %v", err)
	}

	absPath, _ := filepath.Abs(l.ChromiumPath)

	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		ExecutablePath: playwright.String(absPath),
		Headless:       playwright.Bool(false), // Set to false to see it in action!
		Args: []string{
			"--disable-blink-features=AutomationControlled",
		},
	})
	if err != nil {
		log.Fatalf("could not launch CloakBrowser: %v", err)
	}
	defer browser.Close()

	page, err := browser.NewPage()
	if err != nil {
		log.Fatalf("could not create page: %v", err)
	}

	url := l.SearchParams.BuildURL()

	if _, err = page.Goto(url); err != nil {
		log.Fatalf("could not goto: %v", err)
	}

	element := page.Locator("#base-contextual-sign-in-modal > div > section > button > icon > svg")

	if _, err := element.IsVisible(); err == nil {
		element.Click()
	}

	listSelector := "#main-content > section.two-pane-serp-page__results-list > ul > li"

	entries := page.Locator(listSelector)

	count, err := entries.Count()
	if err != nil {
		log.Fatalf("could not count entries: %v", err)
	}

	for i := range count {
		sleep_time := rand.Intn(5)
		time.Sleep(time.Duration(sleep_time) * time.Second)
		entry := entries.Nth(i)

		entry.Click()

		title, _ := entry.Locator("h3").InnerText()

		company, _ := entry.Locator("h4").InnerText()

		location, _ := entry.Locator(".job-search-card__location").InnerText()

		company_image_url, _ := entry.Locator(".artdeco-entity-image").GetAttribute("src")

		description, _ := page.Locator(".show-more-less-html__markup").InnerText()
		job_url, _ := entry.Locator(".base-card__full-link").GetAttribute("href")

		job_pattern := regexp.MustCompile(`\?.*`)

		job_url = job_pattern.ReplaceAllString(job_url, "")

		job_flags, _ := page.Locator(".description__job-criteria-item").All()
		job_type := ""
		seniority := ""

		for _, flag := range job_flags {
			flag_type, _ := flag.Locator("h3").InnerText()
			if strings.Contains(flag_type, "Employment type") {
				job_type, _ = flag.Locator("span").InnerText()
				job_type = strings.TrimSpace(job_type)
			}
			if strings.Contains(flag_type, "Seniority level") {
				seniority, _ = flag.Locator("span").InnerText()
				seniority = strings.TrimSpace(seniority)
			}
		}
		metadata, err := json.Marshal(map[string]string{
			"seniority":         seniority,
			"description":       description,
			"company_image_url": company_image_url,
		})
		if err != nil {
			log.Errorf("could not marshal metadata: %v", err)
		}
		pgTimestamptz := pgtype.Timestamptz{
			Time:  time.Now(),
			Valid: true,
		}

		jobs = append(jobs, db.Job{
			SourceUrl:       url,
			SourceName:      "LinkedIn",
			FirstSeen:       pgTimestamptz,
			ApplicationLink: job_url,
			Company:         &company,
			RoleTitle:       &title,
			Locations:       []string{location},
			JobType:         &job_type,
			Metadata:        metadata,
		})

		fmt.Println(job_url)
	}

	fmt.Scanf("%s")

	return jobs
}
