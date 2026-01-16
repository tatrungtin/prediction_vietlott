package scraper

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/tool_predict/api/vietlott"
	"github.com/tool_predict/internal/application/port"
	"github.com/tool_predict/internal/domain/entity"
	"github.com/tool_predict/internal/domain/valueobject"
	"github.com/tool_predict/internal/infrastructure/logger"
	"go.uber.org/zap"
)

// VietlottWebScraper scrapes Vietlott data from their website using goquery
type VietlottWebScraper struct {
	client      *http.Client
	baseURL     string
	timeout     time.Duration
	retryCount  int
	rateLimit   time.Duration
	mu          sync.Mutex
	lastRequest time.Time
}

// NewVietlottWebScraper creates a new Vietlott web scraper
func NewVietlottWebScraper(baseURL string, timeout time.Duration, retryCount int, rateLimit int) *VietlottWebScraper {
	return &VietlottWebScraper{
		client: &http.Client{
			Timeout: timeout,
		},
		baseURL:    baseURL,
		timeout:    timeout,
		retryCount: retryCount,
		rateLimit:  time.Duration(rateLimit) * time.Second,
	}
}

// FetchLatestDraws fetches the most recent draws for a game type
func (s *VietlottWebScraper) FetchLatestDraws(
	ctx context.Context,
	gameType valueobject.GameType,
	limit int,
) ([]*entity.Draw, error) {
	s.waitForRateLimit()

	// Get the results page path
	gameTypeStr := strings.ToLower(string(gameType))
	resultsPath, ok := vietlott.GameTypePathMap[gameTypeStr]
	if !ok {
		return nil, fmt.Errorf("unknown game type: %s", gameType)
	}

	url := s.baseURL + resultsPath

	// Fetch and parse the page
	draws, err := s.scrapeDrawsPage(ctx, gameType, url, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to scrape draws page: %w", err)
	}

	// Limit results
	if len(draws) > limit {
		draws = draws[:limit]
	}

	return draws, nil
}

// FetchAllDraws fetches all draws from a specified date onwards
func (s *VietlottWebScraper) FetchAllDraws(
	ctx context.Context,
	gameType valueobject.GameType,
	fromDate time.Time,
) ([]*entity.Draw, error) {
	s.waitForRateLimit()

	// Fetch a large number of recent draws
	draws, err := s.FetchLatestDraws(ctx, gameType, 1000)
	if err != nil {
		return nil, err
	}

	// Filter by date
	filteredDraws := make([]*entity.Draw, 0)
	for _, draw := range draws {
		if draw.DrawDate.After(fromDate) || draw.DrawDate.Equal(fromDate) {
			filteredDraws = append(filteredDraws, draw)
		}
	}

	return filteredDraws, nil
}

// FetchDrawByNumber fetches a specific draw by its draw number
func (s *VietlottWebScraper) FetchDrawByNumber(
	ctx context.Context,
	gameType valueobject.GameType,
	drawNumber int,
) (*entity.Draw, error) {
	s.waitForRateLimit()

	// Fetch latest and search
	draws, err := s.FetchLatestDraws(ctx, gameType, 100)
	if err != nil {
		return nil, err
	}

	for _, draw := range draws {
		if draw.DrawNumber == drawNumber {
			return draw, nil
		}
	}

	return nil, fmt.Errorf("draw number %d not found for game type %s", drawNumber, gameType)
}

// FetchDrawsByDateRange fetches all draws within a date range
func (s *VietlottWebScraper) FetchDrawsByDateRange(
	ctx context.Context,
	gameType valueobject.GameType,
	startDate time.Time,
	endDate time.Time,
) ([]*entity.Draw, error) {
	s.waitForRateLimit()

	// Fetch all draws and filter
	draws, err := s.FetchAllDraws(ctx, gameType, startDate)
	if err != nil {
		return nil, err
	}

	filteredDraws := make([]*entity.Draw, 0)
	for _, draw := range draws {
		if (draw.DrawDate.After(startDate) || draw.DrawDate.Equal(startDate)) &&
			draw.DrawDate.Before(endDate) {
			filteredDraws = append(filteredDraws, draw)
		}
	}

	return filteredDraws, nil
}

// GetLatestDrawNumber returns the most recent draw number
func (s *VietlottWebScraper) GetLatestDrawNumber(
	ctx context.Context,
	gameType valueobject.GameType,
) (int, error) {
	draws, err := s.FetchLatestDraws(ctx, gameType, 1)
	if err != nil {
		return 0, err
	}

	if len(draws) == 0 {
		return 0, fmt.Errorf("no draws found for game type %s", gameType)
	}

	return draws[0].DrawNumber, nil
}

// scrapeDrawsPage scrapes the draws page and extracts draw data
func (s *VietlottWebScraper) scrapeDrawsPage(
	ctx context.Context,
	gameType valueobject.GameType,
	url string,
	limit int,
) ([]*entity.Draw, error) {
	// Make HTTP request with retry
	var html string
	for attempt := 0; attempt < s.retryCount; attempt++ {
		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml")
		req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; tool_predict/1.0)")

		resp, err := s.client.Do(req)
		if err != nil {
			if attempt < s.retryCount-1 {
				select {
				case <-ctx.Done():
					return nil, ctx.Err()
				case <-time.After(time.Second * time.Duration(attempt+1)):
				}
				continue
			}
			return nil, fmt.Errorf("failed to fetch page after %d attempts: %w", s.retryCount, err)
		}

		if resp.StatusCode == http.StatusOK {
			body, err := io.ReadAll(resp.Body)
			resp.Body.Close()
			if err != nil {
				return nil, fmt.Errorf("failed to read response body: %w", err)
			}
			html = string(body)
			break
		}

		resp.Body.Close()

		if attempt < s.retryCount-1 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(time.Second * time.Duration(attempt+1)):
			}
		} else {
			return nil, fmt.Errorf("server returned status %d", resp.StatusCode)
		}
	}

	// Parse HTML
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	// Extract draw data
	// Note: The actual selectors will depend on Vietlott's website structure
	// These are example selectors that may need adjustment
	draws := make([]*entity.Draw, 0)

	// Common pattern: look for tables or divs containing draw results
	// This is a generic implementation - adjust selectors based on actual HTML
	doc.Find(".result-row, tr.draw-row, .lottery-result").Each(func(i int, row *goquery.Selection) {
		if len(draws) >= limit {
			return
		}

		draw, err := s.parseDrawRow(gameType, row)
		if err != nil {
			logger.Warn("Failed to parse draw row",
				zap.Int("row", i),
				zap.Error(err),
			)
			return
		}

		draws = append(draws, draw)
	})

	if len(draws) == 0 {
		return nil, fmt.Errorf("no draws found on page")
	}

	return draws, nil
}

// parseDrawRow parses a single draw row from HTML
func (s *VietlottWebScraper) parseDrawRow(gameType valueobject.GameType, sel *goquery.Selection) (*entity.Draw, error) {
	// Extract draw number
	drawNumberText := sel.Find(".draw-number, .period, .ky").First().Text()
	drawNumberText = strings.TrimSpace(drawNumberText)
	drawNumber, err := strconv.Atoi(drawNumberText)
	if err != nil {
		return nil, fmt.Errorf("failed to parse draw number: %w", err)
	}

	// Extract numbers
	var numbers []int
	sel.Find(".number, .ball, .lottery-num").Each(func(i int, numSel *goquery.Selection) {
		if len(numbers) >= 6 {
			return
		}
		numText := numSel.Text()
		numText = strings.TrimSpace(numText)
		num, err := strconv.Atoi(numText)
		if err == nil {
			numbers = append(numbers, num)
		}
	})

	if len(numbers) != 6 {
		return nil, fmt.Errorf("expected 6 numbers, got %d", len(numbers))
	}

	numbersVO, err := valueobject.NewNumbers(numbers)
	if err != nil {
		return nil, fmt.Errorf("invalid numbers: %w", err)
	}

	// Extract date
	dateText := sel.Find(".draw-date, .date, .ngay").First().Text()
	dateText = strings.TrimSpace(dateText)
	drawDate, err := time.Parse("02/01/2006", dateText) // DD/MM/YYYY format
	if err != nil {
		// Try alternative formats
		drawDate, err = time.Parse("2006-01-02", dateText)
		if err != nil {
			return nil, fmt.Errorf("failed to parse date: %w", err)
		}
	}

	// Extract jackpot (optional)
	jackpotText := sel.Find(".jackpot, .prize").First().Text()
	jackpotText = strings.TrimSpace(jackpotText)
	jackpotText = strings.ReplaceAll(jackpotText, ",", "")
	jackpotText = strings.ReplaceAll(jackpotText, ".", "")
	jackpot, _ := strconv.ParseFloat(jackpotText, 64)

	// Extract winners (optional)
	winnersText := sel.Find(".winners, .winner-count").First().Text()
	winnersText = strings.TrimSpace(winnersText)
	winners, _ := strconv.Atoi(winnersText)

	// Create draw entity
	draw, err := entity.NewDraw(
		gameType,
		drawNumber,
		numbersVO,
		drawDate,
		jackpot,
		winners,
	)

	return draw, err
}

// waitForRateLimit implements rate limiting
func (s *VietlottWebScraper) waitForRateLimit() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.rateLimit > 0 {
		timeSinceLastRequest := time.Since(s.lastRequest)
		if timeSinceLastRequest < s.rateLimit {
			time.Sleep(s.rateLimit - timeSinceLastRequest)
		}
		s.lastRequest = time.Now()
	}
}

// Ensure VietlottWebScraper implements port.VietlottScraper
var _ port.VietlottScraper = (*VietlottWebScraper)(nil)
