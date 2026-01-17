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
	winningNumberURL = "https://vietlott.vn/vi/trung-thuong/ket-qua-trung-thuong/winning-number-645"
	outputDir        = "data/draws/mega_6_45"
	targetDays       = 30 // Target number of draws
	gameType         = "MEGA_6_45"
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

	fmt.Printf("🎲 Crawling Vietlott Mega 6/45...\n\n")

	// Fetch latest draws from the winning number page
	draws, err := fetchLatestDraws()
	if err != nil {
		log.Fatalf("Failed to fetch draws: %v", err)
	}

	if len(draws) == 0 {
		fmt.Println("No draws found on Vietlott website")
		return
	}

	fmt.Printf("Found %d draws on Vietlott website\n\n", len(draws))

	// Check what we already have
	existingDraws := getExistingDraws()
	newDraws := 0

	// Save draws
	for _, draw := range draws {
		if !existingDraws[draw.DrawNumber] {
			if err := saveDraw(draw); err != nil {
				log.Printf("Error saving draw %d: %v", draw.DrawNumber, err)
				continue
			}
			newDraws++
			fmt.Printf("  ✓ New: %s - %s - Numbers: %v\n",
				draw.ID,
				draw.DrawDate.Format("02/01/2006"),
				draw.Numbers)
		}
	}

	fmt.Printf("\n✅ Saved %d new draws\n", newDraws)

	// Show totals
	totalDraws := len(existingDraws) + newDraws
	fmt.Printf("📊 Total draws: %d/30\n", totalDraws)

	if totalDraws < targetDays {
		fmt.Printf("⏳ Need %d more draws (will accumulate over time via daily crawler)\n", targetDays-totalDraws)
	}

	fmt.Printf("\n📅 Daily GitHub Actions will fetch new draws automatically\n")
}

func getExistingDraws() map[int]bool {
	draws := make(map[int]bool)

	entries, err := os.ReadDir(outputDir)
	if err != nil {
		return draws
	}

	re := regexp.MustCompile(`mega_(\d+)\.json`)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		matches := re.FindStringSubmatch(entry.Name())
		if len(matches) < 2 {
			continue
		}

		drawNum, err := strconv.Atoi(matches[1])
		if err != nil {
			continue
		}

		draws[drawNum] = true
	}

	return draws
}

func fetchLatestDraws() ([]Draw, error) {
	req, err := http.NewRequest("GET", winningNumberURL, nil)
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

		// Parse draw number from URL (e.g., "/vi/trung-thuong/ket-qua-trung-thuong/645?id=01295&nocatche=1")
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
			if err == nil && num <= 45 { // Only include numbers 01-45 for Mega 6/45
				numbers = append(numbers, num)
			}
		})

		// We need exactly 6 main numbers
		if len(numbers) < 6 {
			return
		}

		// Parse date
		drawDate, err := time.Parse(dateLayout, dateStr)
		if err != nil {
			return
		}

		draw := Draw{
			ID:         fmt.Sprintf("mega_%05d", drawNumber),
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
	filePath := fmt.Sprintf("%s/mega_%05d.json", outputDir, draw.DrawNumber)

	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")

	return encoder.Encode(draw)
}
