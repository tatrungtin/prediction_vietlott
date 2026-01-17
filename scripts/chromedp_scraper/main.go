package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/chromedp"
)

const (
	announcementURL = "https://vietlott.vn/vi/trung-thuong/ket-qua-trung-thuong/thong-bao-ket-qua-655"
	detailURLBase   = "https://vietlott.vn/vi/trung-thuong/ket-qua-trung-thuong/655?id=%s"
	outputDir       = "data/draws/power_6_55"
	gameType        = "POWER_6_55"
	targetPages     = 5 // Number of pages to crawl from announcement page
)

// Draw represents a lottery draw
type Draw struct {
	ID         string    `json:"id"`
	GameType   string    `json:"game_type"`
	DrawNumber int       `json:"draw_number"`
	Numbers    []int     `json:"numbers"`
	DrawDate   time.Time `json:"draw_date"`
	Jackpot    float64   `json:"jackpot"`
	Winners    int       `json:"winners"`
}

func main() {
	log.Println("Starting Vietlott Power 6/55 crawler with headless browser...")

	// Ensure output directory exists
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		log.Fatalf("Failed to create output directory: %v", err)
	}

	// Get existing draw numbers
	existingDraws := getExistingDraws()
	log.Printf("Found %d existing draws", len(existingDraws))

	// Crawl draws from announcement pages using headless browser
	draws, err := crawlFromAnnouncementPages()
	if err != nil {
		log.Fatalf("Failed to crawl from announcement pages: %v", err)
	}

	log.Printf("Crawled %d draws from announcement pages", len(draws))

	// Save draws that don't already exist
	savedCount := 0
	for _, draw := range draws {
		drawID := fmt.Sprintf("power_%05d", draw.DrawNumber)
		if _, exists := existingDraws[drawID]; !exists {
			if err := saveDraw(draw); err != nil {
				log.Printf("Failed to save draw %d: %v", draw.DrawNumber, err)
			} else {
				savedCount++
			}
		}
	}

	log.Printf("Saved %d new draws (skipped %d duplicates)", savedCount, len(draws)-savedCount)
	log.Println("Crawl completed!")
}

// getExistingDraws returns a map of existing draw IDs
func getExistingDraws() map[string]bool {
	existing := make(map[string]bool)

	files, err := os.ReadDir(outputDir)
	if err != nil {
		return existing
	}

	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".json") {
			drawID := strings.TrimSuffix(file.Name(), ".json")
			existing[drawID] = true
		}
	}

	return existing
}

// crawlFromAnnouncementPages crawls draws from announcement pages using headless browser
func crawlFromAnnouncementPages() ([]*Draw, error) {
	// Create context with options to bypass sandbox restrictions on CI/CD
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.Flag("disable-gpu", true),
	)

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	// Set timeout
	ctx, cancel = context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	allDraws := make([]*Draw, 0)

	// Iterate through pages
	for page := 1; page <= targetPages; page++ {
		log.Printf("Crawling announcement page %d/%d...", page, targetPages)

		// Get draws from this page
		draws, err := getDrawsFromAnnouncementPage(ctx, page)
		if err != nil {
			log.Printf("Error getting draws from page %d: %v", page, err)
			continue
		}

		log.Printf("Found %d draws on page %d", len(draws), page)
		allDraws = append(allDraws, draws...)

		// Don't rate limit too much between pages
		time.Sleep(2 * time.Second)
	}

	// Fetch winning numbers for each draw using headless browser
	log.Printf("Fetching winning numbers for %d draws...", len(allDraws))

	for i, draw := range allDraws {
		if draw.Numbers == nil || len(draw.Numbers) == 0 {
			log.Printf("[%d/%d] Fetching numbers for draw %d...", i+1, len(allDraws), draw.DrawNumber)

			numbers, err := fetchDrawNumbersWithBrowser(ctx, draw.DrawNumber)
			if err != nil {
				log.Printf("Failed to fetch numbers for draw %d: %v", draw.DrawNumber, err)
				continue
			}

			draw.Numbers = numbers
			log.Printf("Draw %d: %v", draw.DrawNumber, numbers)
		}

		// Small delay between requests
		time.Sleep(1 * time.Second)
	}

	return allDraws, nil
}

// getDrawsFromAnnouncementPage gets draw information from an announcement page
func getDrawsFromAnnouncementPage(ctx context.Context, pageNum int) ([]*Draw, error) {
	var htmlContent string

	// Navigate to the announcement page
	err := chromedp.Run(ctx,
		chromedp.Navigate(announcementURL),
		chromedp.Sleep(2*time.Second),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to navigate to announcement page: %w", err)
	}

	// If not the first page, click on pagination
	if pageNum > 1 {
		// Find and click the page link (li index = pageNum + 1)
		pageSelector := fmt.Sprintf(`ul.pagination li:nth-child(%d) a`, pageNum+1)
		err := chromedp.Run(ctx,
			chromedp.Click(pageSelector, chromedp.ByQuery),
			chromedp.Sleep(2*time.Second),
		)
		if err != nil {
			// Try alternative method if click fails
			log.Printf("Warning: Could not click page %d with selector %s: %v", pageNum, pageSelector, err)
			// Continue with current page content
		}
	}

	// Get the full body HTML after navigation
	err = chromedp.Run(ctx,
		chromedp.OuterHTML("body", &htmlContent, chromedp.ByQuery),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get HTML: %w", err)
	}

	// Parse HTML
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	draws := make([]*Draw, 0)

	// Extract draw information from table rows
	doc.Find("table tbody tr").Each(func(i int, s *goquery.Selection) {
		// Skip header row
		if i == 0 {
			return
		}

		// Get the "Chi tiết" link (PDF link)
		pdfLink := s.Find("a").FilterFunction(func(i int, s *goquery.Selection) bool {
			href, exists := s.Attr("href")
			return exists && strings.Contains(href, ".pdf")
		})

		if pdfLink.Length() == 0 {
			return
		}

		href, _ := pdfLink.Attr("href")

		// Extract draw number from PDF filename
		// Format: https://media.vietlott.vn/.../26.01.15---[655]---01295---draw-result.pdf
		re := regexp.MustCompile(`\[655\]---(\d+)---`)
		matches := re.FindStringSubmatch(href)
		if len(matches) < 2 {
			return
		}

		drawNumber, err := strconv.Atoi(matches[1])
		if err != nil {
			return
		}

		// Get text from the row to extract date
		rowText := s.Text()

		// Parse date from text (format: "POWER 6/55 - Kỳ 00687 - Ngày 19/02/2022")
		dateRe := regexp.MustCompile(`Ngày (\d{2})/(\d{2})/(\d{4})`)
		dateMatches := dateRe.FindStringSubmatch(rowText)
		if len(dateMatches) < 4 {
			return
		}

		day, _ := strconv.Atoi(dateMatches[1])
		month, _ := strconv.Atoi(dateMatches[2])
		year, _ := strconv.Atoi(dateMatches[3])

		drawDate := time.Date(year, time.Month(month), day, 18, 0, 0, 0, time.UTC)

		draw := &Draw{
			ID:         fmt.Sprintf("power_%05d", drawNumber),
			GameType:   gameType,
			DrawNumber: drawNumber,
			DrawDate:   drawDate,
		}

		draws = append(draws, draw)
	})

	return draws, nil
}

// fetchDrawNumbersWithBrowser fetches winning numbers for a specific draw using headless browser
func fetchDrawNumbersWithBrowser(ctx context.Context, drawNumber int) ([]int, error) {
	url := fmt.Sprintf(detailURLBase, fmt.Sprintf("%05d", drawNumber))

	var htmlContent string
	var numbersText string

	// Navigate to detail page and extract numbers
	err := chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.WaitReady(`body`, chromedp.ByQuery),
		// Wait for JavaScript to load the numbers
		chromedp.Sleep(2*time.Second),
		chromedp.Text(`.result`, &numbersText, chromedp.ByQuery),
		chromedp.OuterHTML(`body`, &htmlContent, chromedp.ByQuery),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to navigate to detail page: %w", err)
	}

	// Try multiple strategies to extract numbers
	numbers := extractNumbersFromHTML(htmlContent)
	if len(numbers) == 6 {
		return numbers, nil
	}

	// If standard extraction fails, try alternative methods
	numbers = extractNumbersAlternative(htmlContent)
	if len(numbers) == 6 {
		return numbers, nil
	}

	return nil, fmt.Errorf("could not extract 6 numbers from page")
}

// extractNumbersFromHTML extracts winning numbers from HTML
func extractNumbersFromHTML(html string) []int {
	// Try to find numbers in common patterns
	patterns := []string{
		`(\d{1,2})-(\d{1,2})-(\d{1,2})-(\d{1,2})-(\d{1,2})-(\d{1,2})`,
		`(\d{1,2})\s*-\s*(\d{1,2})\s*-\s*(\d{1,2})\s*-\s*(\d{1,2})\s*-\s*(\d{1,2})\s*-\s*(\d{1,2})`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(html)
		if len(matches) == 7 {
			numbers := make([]int, 6)
			for i := 0; i < 6; i++ {
				num, _ := strconv.Atoi(matches[i+1])
				numbers[i] = num
			}
			return numbers
		}
	}

	return []int{}
}

// extractNumbersAlternative tries alternative methods to extract numbers
func extractNumbersAlternative(html string) []int {
	// Look for numbers in ball/result elements
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return []int{}
	}

	numbers := make([]int, 0)

	// Try finding numbers in various CSS selectors
	selectors := []string{
		".ball",
		".number",
		".result-number",
		"[class*='ball']",
		"[class*='number']",
	}

	for _, selector := range selectors {
		doc.Find(selector).Each(func(i int, s *goquery.Selection) {
			text := strings.TrimSpace(s.Text())
			num, err := strconv.Atoi(text)
			if err == nil && num > 0 && num <= 55 {
				numbers = append(numbers, num)
			}
		})

		if len(numbers) == 6 {
			return numbers
		}

		numbers = numbers[:0]
	}

	return numbers
}

// saveDraw saves a draw to a JSON file
func saveDraw(draw *Draw) error {
	data, err := json.MarshalIndent(draw, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal draw: %w", err)
	}

	filename := fmt.Sprintf("power_%05d.json", draw.DrawNumber)
	filepath := filepath.Join(outputDir, filename)

	if err := os.WriteFile(filepath, data, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// fetchDrawNumbersWithoutBrowser attempts to fetch numbers without using browser
// (fallback method using HTTP requests)
func fetchDrawNumbersWithoutBrowser(drawNumber int) ([]int, error) {
	url := fmt.Sprintf(detailURLBase, fmt.Sprintf("%05d", drawNumber))

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	// Try to extract numbers from the page
	var numbers []int

	doc.Find(".ball, .number").Each(func(i int, s *goquery.Selection) {
		text := strings.TrimSpace(s.Text())
		num, err := strconv.Atoi(text)
		if err == nil && num > 0 && num <= 55 {
			numbers = append(numbers, num)
		}
	})

	if len(numbers) == 6 {
		return numbers, nil
	}

	return nil, fmt.Errorf("could not extract 6 numbers from page")
}
