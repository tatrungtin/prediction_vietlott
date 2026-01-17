package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/chromedp"
)

const (
	baseURL        = "https://vietlott.vn/vi/trung-thuong/ket-qua-trung-thuong/thong-bao-ket-qua-655?pageindex=%d&nocatche=1"
	outputDir      = "data/draws/power_6_55"
	gameType       = "POWER_6_55"
	totalPages     = 5
	pdfDownloadDir = "/tmp/vietlott_pdfs"
	imageOutputDir = "/tmp/vietlott_images"
)

type Draw struct {
	ID         string    `json:"id"`
	GameType   string    `json:"game_type"`
	DrawNumber int       `json:"draw_number"`
	Numbers    []int     `json:"numbers"`
	DrawDate   time.Time `json:"draw_date"`
}

func main() {
	log.Println("Starting MCP-based OCR crawler for Vietlott Power 6/55...")
	log.Println("This will:")
	log.Println("1. Download PDFs from announcement pages")
	log.Println("2. Convert PDFs to images")
	log.Println("3. Use MCP AI vision to extract winning numbers")
	log.Println("")

	// Ensure output directory exists
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		log.Fatalf("Failed to create output directory: %v", err)
	}

	// Create temp directories
	if err := os.MkdirAll(pdfDownloadDir, 0755); err != nil {
		log.Fatalf("Failed to create PDF directory: %v", err)
	}
	if err := os.MkdirAll(imageOutputDir, 0755); err != nil {
		log.Fatalf("Failed to create image directory: %v", err)
	}

	// Get existing draws
	existingDraws := getExistingDraws()
	log.Printf("Found %d existing draws", len(existingDraws))

	allDraws := make([]*Draw, 0)

	// Crawl each page
	for page := 1; page <= totalPages; page++ {
		log.Printf("=== Crawling page %d/%d ===", page, totalPages)

		url := fmt.Sprintf(baseURL, page)
		draws, err := crawlPageWithMCPOCR(url, page)
		if err != nil {
			log.Printf("Error crawling page %d: %v", page, err)
			continue
		}

		log.Printf("Found %d draws on page %d", len(draws), page)
		allDraws = append(allDraws, draws...)
		time.Sleep(2 * time.Second)
	}

	log.Printf("\n=== Total draws found: %d ===", len(allDraws))

	// Save new draws
	savedCount := 0
	skippedCount := 0

	for _, draw := range allDraws {
		drawID := fmt.Sprintf("power_%05d", draw.DrawNumber)
		if _, exists := existingDraws[drawID]; exists {
			skippedCount++
			continue
		}

		if err := saveDraw(draw); err != nil {
			log.Printf("Failed to save draw %d: %v", draw.DrawNumber, err)
		} else {
			savedCount++
			log.Printf("✓ Saved draw %05d: %v (%s)", draw.DrawNumber, draw.Numbers, draw.DrawDate.Format("2006-01-02"))
		}
	}

	log.Printf("\n=== Final Summary ===")
	log.Printf("Saved: %d new draws", savedCount)
	log.Printf("Skipped: %d existing draws", skippedCount)
	log.Printf("Total: %d draws processed", len(allDraws))
}

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

func crawlPageWithMCPOCR(url string, pageNum int) ([]*Draw, error) {
	// Use chromedp to fetch the page
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	var htmlContent string
	err := chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.Sleep(3*time.Second),
		chromedp.OuterHTML("body", &htmlContent, chromedp.ByQuery),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch page: %w", err)
	}

	// Parse HTML
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	draws := make([]*Draw, 0)

	// Extract draws from table rows
	doc.Find("table tbody tr").Each(func(i int, s *goquery.Selection) {
		if i == 0 {
			return // Skip header
		}

		// Get PDF link
		pdfLink := s.Find("a").FilterFunction(func(i int, s *goquery.Selection) bool {
			href, exists := s.Attr("href")
			return exists && strings.Contains(href, ".pdf")
		})

		if pdfLink.Length() == 0 {
			return
		}

		href, _ := pdfLink.Attr("href")

		// Extract draw number from PDF URL
		// Format: https://media.vietlott.vn/.../25.12.23---[655]---01285---draw-result.pdf
		re := regexp.MustCompile(`\[655\]---(\d+)---`)
		matches := re.FindStringSubmatch(href)
		if len(matches) < 2 {
			return
		}

		drawNumber, err := strconv.Atoi(matches[1])
		if err != nil {
			return
		}

		// Extract date from row text
		rowText := s.Text()
		dateRe := regexp.MustCompile(`Ngày (\d{2})/(\d{2})/(\d{4})`)
		dateMatches := dateRe.FindStringSubmatch(rowText)
		if len(dateMatches) < 4 {
			return
		}

		day, _ := strconv.Atoi(dateMatches[1])
		month, _ := strconv.Atoi(dateMatches[2])
		year, _ := strconv.Atoi(dateMatches[3])
		drawDate := time.Date(year, time.Month(month), day, 18, 0, 0, 0, time.UTC)

		// Try to extract numbers from PDF using MCP OCR
		numbers, err := extractNumbersWithMCPOCR(href, drawNumber)
		if err != nil {
			log.Printf("  Warning: Could not extract numbers for draw %d: %v", drawNumber, err)
			return
		}

		if len(numbers) != 6 {
			log.Printf("  Warning: Expected 6 numbers for draw %d, got %d: %v", drawNumber, len(numbers), numbers)
			return
		}

		draw := &Draw{
			ID:         fmt.Sprintf("power_%05d", drawNumber),
			GameType:   gameType,
			DrawNumber: drawNumber,
			Numbers:    numbers,
			DrawDate:   drawDate,
		}

		draws = append(draws, draw)
		log.Printf("  ✓ Draw %05d: %v", drawNumber, numbers)
	})

	return draws, nil
}

func extractNumbersWithMCPOCR(pdfURL string, drawNumber int) ([]int, error) {
	// Download PDF
	pdfFilename := fmt.Sprintf("draw_%05d.pdf", drawNumber)
	pdfPath := filepath.Join(pdfDownloadDir, pdfFilename)

	// Check if we already downloaded this PDF
	if _, err := os.Stat(pdfPath); os.IsNotExist(err) {
		log.Printf("  [%05d] Downloading PDF...", drawNumber)

		// Download the PDF
		client := &http.Client{
			Timeout: 30 * time.Second,
		}

		req, err := http.NewRequest("GET", pdfURL, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; tool_predict/1.0)")

		resp, err := client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("failed to download PDF: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("PDF download returned status %d", resp.StatusCode)
		}

		// Save PDF to file
		outFile, err := os.Create(pdfPath)
		if err != nil {
			return nil, fmt.Errorf("failed to create PDF file: %w", err)
		}
		defer outFile.Close()

		_, err = outFile.ReadFrom(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to write PDF file: %w", err)
		}
	}

	// Convert PDF to image
	imagePath, err := convertPDFToImage(pdfPath, drawNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to convert PDF to image: %w", err)
	}

	// Perform OCR on the image using MCP AI Vision
	text, err := performMCPOCR(imagePath)
	if err != nil {
		return nil, fmt.Errorf("failed to perform MCP OCR: %w", err)
	}

	// Extract 6 numbers from OCR text
	numbers := extractNumbersFromText(text)

	if len(numbers) != 6 {
		// Log OCR text for debugging
		log.Printf("  [%05d] OCR Text: %s", drawNumber, text)
	}

	return numbers, nil
}

func convertPDFToImage(pdfPath string, drawNumber int) (string, error) {
	// Create image output path
	imageBase := filepath.Join(imageOutputDir, fmt.Sprintf("draw_%05d", drawNumber))

	// Check if image already exists
	imagePath := imageBase + ".png"
	if _, err := os.Stat(imagePath); err == nil {
		// Image already exists from previous run
		return imagePath, nil
	}

	// Use pdftoppm to convert PDF to PNG image
	cmd := exec.Command("pdftoppm", "-png", "-singlefile", pdfPath, imageBase)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("pdftoppm failed: %w, output: %s", err, string(output))
	}

	// pdftoppm creates a file with .png extension
	if _, err := os.Stat(imagePath); os.IsNotExist(err) {
		return "", fmt.Errorf("image file not created: %s", imagePath)
	}

	return imagePath, nil
}

func performMCPOCR(imagePath string) (string, error) {
	// Since we can't directly call MCP tools from Go code,
	// we'll write a Python script that uses the MCP vision model
	// For now, let's use a simpler approach with tesseract if available

	// Create a Python script that would call MCP if it were integrated
	// For now, we'll save the image path for manual processing
	return "", fmt.Errorf("MCP OCR needs to be implemented via Python script - image saved at %s", imagePath)
}

func extractNumbersFromText(text string) []int {
	// Clean up the text
	text = strings.ReplaceAll(text, "|", "1")
	text = strings.ReplaceAll(text, "l", "1")
	text = strings.ReplaceAll(text, "O", "0")
	text = strings.ReplaceAll(text, "o", "0")
	text = strings.ReplaceAll(text, "S", "5")
	text = strings.ReplaceAll(text, "Z", "2")

	// Try multiple patterns
	patterns := []string{
		`(\d{1,2})\s*[-–—]\s*(\d{1,2})\s*[-–—]\s*(\d{1,2})\s*[-–—]\s*(\d{1,2})\s*[-–—]\s*(\d{1,2})\s*[-–—]\s*(\d{1,2})`,
		`(\d{1,2})\s+(\d{1,2})\s+(\d{1,2})\s+(\d{1,2})\s+(\d{1,2})\s+(\d{1,2})`,
		`Bộ\s*số\s*[:：]?\s*(\d{1,2}).*?(\d{1,2}).*?(\d{1,2}).*?(\d{1,2}).*?(\d{1,2}).*?(\d{1,2})`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(text)
		if len(matches) >= 7 {
			numbers := make([]int, 6)
			valid := true
			for i := 0; i < 6; i++ {
				num, err := strconv.Atoi(strings.TrimSpace(matches[i+1]))
				if err != nil || num < 1 || num > 55 {
					valid = false
					break
				}
				numbers[i] = num
			}
			if valid {
				return numbers
			}
		}
	}

	// Fallback
	re := regexp.MustCompile(`\b(\d{1,2})\b`)
	matches := re.FindAllStringSubmatch(text, -1)

	numbers := make([]int, 0)
	for _, match := range matches {
		if len(match) > 1 {
			num, err := strconv.Atoi(match[1])
			if err == nil && num >= 1 && num <= 55 {
				duplicate := false
				for _, existing := range numbers {
					if existing == num {
						duplicate = true
						break
					}
				}
				if !duplicate {
					numbers = append(numbers, num)
				}
			}
		}
	}

	if len(numbers) > 6 {
		for i := 0; i <= len(numbers)-6; i++ {
			candidate := numbers[i : i+6]
			isSorted := true
			for j := 1; j < 6; j++ {
				if candidate[j] <= candidate[j-1] {
					isSorted = false
					break
				}
			}
			if isSorted {
				return candidate
			}
		}
		return numbers[len(numbers)-6:]
	}

	if len(numbers) == 6 {
		return numbers
	}

	return []int{}
}

func saveDraw(draw *Draw) error {
	data, err := json.MarshalIndent(draw, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal draw: %w", err)
	}

	filename := fmt.Sprintf("power_%05d.json", draw.DrawNumber)
	filepath := fmt.Sprintf("%s/%s", outputDir, filename)

	return os.WriteFile(filepath, data, 0644)
}
