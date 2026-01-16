package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

const (
	winningNumberURL = "https://vietlott.vn/vi/trung-thuong/ket-qua-trung-thuong/winning-number-655"
	outputDir        = "data/draws/power_6_55"
	targetDays       = 30   // Number of days to fetch
	maxPages         = 5    // Maximum pages to try
	gameType         = "POWER_6_55"
	dateLayout       = "02/01/2006"
)

type Draw struct {
	ID         string    `json:"id"`
	GameType   string    `json:"game_type"`
	DrawNumber int       `json:"draw_number"`
	Numbers    []int     `json:"numbers"`
	DrawDate   time.Time `json:"draw_date"`
	Jackpot    int       `json:"jackpot"`
	Winners    int       `json:"winners"`
}

func main() {
	// Create output directory
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		log.Fatalf("Failed to create directory: %v", err)
	}

	// Fetch draws from the winning number page
	draws := make([]Draw, 0)

	for page := 0; page < maxPages; page++ {
		fmt.Printf("Fetching page %d...\n", page+1)

		url := winningNumberURL
		if page > 0 {
			// Try to get page via AJAX (this may not work without JS execution)
			url = fmt.Sprintf("%s?page=%d", winningNumberURL, page)
		}

		pageDraws, err := fetchPage(url)
		if err != nil {
			log.Printf("Error fetching page %d: %v", page+1, err)
			// If page 0 fails, we can't continue
			if page == 0 {
				log.Fatalf("Failed to fetch first page")
			}
			break
		}

		if len(pageDraws) == 0 {
			fmt.Printf("No draws found on page %d, stopping\n", page+1)
			break
		}

		draws = append(draws, pageDraws...)
		fmt.Printf("  Found %d draws on page %d\n", len(pageDraws), page+1)

		// Check if we have enough data
		if len(draws) >= targetDays {
			fmt.Printf("Reached target of %d draws\n", targetDays)
			draws = draws[:targetDays] // Trim to exactly targetDays
			break
		}

		// Small delay to be polite
		time.Sleep(500 * time.Millisecond)
	}

	fmt.Printf("\nTotal draws fetched: %d\n", len(draws))

	// Save draws to individual JSON files
	savedCount := 0
	for _, draw := range draws {
		if err := saveDraw(draw); err != nil {
			log.Printf("Error saving draw %d: %v", draw.DrawNumber, err)
			continue
		}
		savedCount++
		fmt.Printf("  ✓ Saved: %s - %s - Numbers: %v\n",
			draw.ID,
			draw.DrawDate.Format("02/01/2006"),
			draw.Numbers)
	}

	fmt.Printf("\nSuccessfully saved %d draws to %s\n", savedCount, outputDir)
}

func fetchPage(url string) ([]Draw, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	// Set headers to mimic a real browser
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", "vi-VN,vi;q=0.9,en-US;q=0.8,en;q=0.7")

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status code: %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	draws := make([]Draw, 0)

	// Find the table with draw results
	doc.Find("tbody tr").Each(func(i int, s *goquery.Selection) {
		// Extract date from first column
		dateCell := s.Find("td").First()
		dateStr := strings.TrimSpace(dateCell.Text())

		// Extract draw number from second column
		drawNumberCell := s.Find("td").Eq(1)
		drawNumberLink := drawNumberCell.Find("a")
		href, exists := drawNumberLink.Attr("href")
		if !exists {
			return
		}

		// Parse draw number from URL (e.g., "/vi/trung-thuong/ket-qua-trung-thuong/655?id=01295&nocatche=1")
		re := regexp.MustCompile(`id=(\d+)`)
		matches := re.FindStringSubmatch(href)
		if len(matches) < 2 {
			return
		}

		drawNumber, err := strconv.Atoi(matches[1])
		if err != nil {
			return
		}

		// Extract numbers from the ball display in third column
		numbersDiv := s.Find(".day_so_ket_qua_v2")
		numbers := make([]int, 0)
		numbersDiv.Find(".bong_tron").Each(func(j int, ball *goquery.Selection) {
			// Skip the separator element
			if ball.HasClass("bong_tron-sperator") {
				return
			}
			numStr := strings.TrimSpace(ball.Text())
			num, err := strconv.Atoi(numStr)
			if err == nil {
				numbers = append(numbers, num)
			}
		})

		// We need exactly 6 main numbers (ignore the 7th power number)
		if len(numbers) < 6 {
			return
		}

		// Parse date
		drawDate, err := time.Parse(dateLayout, dateStr)
		if err != nil {
			return
		}

		draw := Draw{
			ID:         fmt.Sprintf("power_%05d", drawNumber),
			GameType:   gameType,
			DrawNumber: drawNumber,
			Numbers:    numbers[:6], // Only take first 6 numbers
			DrawDate:   drawDate,
			Jackpot:    0,
			Winners:    0,
		}

		draws = append(draws, draw)
	})

	return draws, nil
}

func saveDraw(draw Draw) error {
	filePath := fmt.Sprintf("%s/power_%05d.json", outputDir, draw.DrawNumber)

	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")

	return encoder.Encode(draw)
}
