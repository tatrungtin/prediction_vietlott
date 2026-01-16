package scraper

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/tool_predict/api/vietlott"
	"github.com/tool_predict/internal/application/port"
	"github.com/tool_predict/internal/domain/entity"
	"github.com/tool_predict/internal/domain/valueobject"
	"github.com/tool_predict/internal/infrastructure/logger"
	"go.uber.org/zap"
)

// VietlottAPIScraper scrapes Vietlott data using their API (if available)
// Falls back to web scraping if API is not accessible
type VietlottAPIScraper struct {
	client      *http.Client
	baseURL     string
	timeout     time.Duration
	retryCount  int
	rateLimit   time.Duration
	lastRequest time.Time
}

// NewVietlottAPIScraper creates a new Vietlott API scraper
func NewVietlottAPIScraper(baseURL string, timeout time.Duration, retryCount int, rateLimit int) *VietlottAPIScraper {
	return &VietlottAPIScraper{
		client: &http.Client{
			Timeout: timeout,
			Transport: &http.Transport{
				MaxIdleConns:       10,
				IdleConnTimeout:    30 * time.Second,
				DisableCompression: false,
			},
		},
		baseURL:    baseURL,
		timeout:    timeout,
		retryCount: retryCount,
		rateLimit:  time.Duration(rateLimit) * time.Second,
	}
}

// FetchLatestDraws fetches the most recent draws for a game type
func (s *VietlottAPIScraper) FetchLatestDraws(
	ctx context.Context,
	gameType valueobject.GameType,
	limit int,
) ([]*entity.Draw, error) {
	// Rate limiting
	s.waitForRateLimit()

	// Try API first, fall back to web scraping
	draws, err := s.fetchFromAPI(ctx, gameType, limit)
	if err != nil {
		logger.Warn("API fetch failed, falling back to web scraping",
			zap.String("game_type", string(gameType)),
			zap.Error(err),
		)
		// Fall back to web scraper
		webScraper := NewVietlottWebScraper(s.baseURL, s.timeout, s.retryCount, int(s.rateLimit.Seconds()))
		return webScraper.FetchLatestDraws(ctx, gameType, limit)
	}

	return draws, nil
}

// FetchAllDraws fetches all draws from a specified date onwards
func (s *VietlottAPIScraper) FetchAllDraws(
	ctx context.Context,
	gameType valueobject.GameType,
	fromDate time.Time,
) ([]*entity.Draw, error) {
	// Rate limiting
	s.waitForRateLimit()

	// Try API first
	draws, err := s.fetchFromAPI(ctx, gameType, 1000) // Fetch large batch
	if err != nil {
		logger.Warn("API fetch failed, falling back to web scraping",
			zap.String("game_type", string(gameType)),
			zap.Error(err),
		)
		webScraper := NewVietlottWebScraper(s.baseURL, s.timeout, s.retryCount, int(s.rateLimit.Seconds()))
		return webScraper.FetchAllDraws(ctx, gameType, fromDate)
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
func (s *VietlottAPIScraper) FetchDrawByNumber(
	ctx context.Context,
	gameType valueobject.GameType,
	drawNumber int,
) (*entity.Draw, error) {
	// Rate limiting
	s.waitForRateLimit()

	// For now, fetch latest and search
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
func (s *VietlottAPIScraper) FetchDrawsByDateRange(
	ctx context.Context,
	gameType valueobject.GameType,
	startDate time.Time,
	endDate time.Time,
) ([]*entity.Draw, error) {
	// Rate limiting
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
func (s *VietlottAPIScraper) GetLatestDrawNumber(
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

// fetchFromAPI attempts to fetch data from the API
func (s *VietlottAPIScraper) fetchFromAPI(
	ctx context.Context,
	gameType valueobject.GameType,
	limit int,
) ([]*entity.Draw, error) {
	// Construct API URL
	gameTypeStr := strings.ToLower(string(gameType))
	apiPath, ok := vietlott.GameTypePathMap[gameTypeStr]
	if !ok {
		return nil, fmt.Errorf("unknown game type: %s", gameType)
	}

	u, err := url.Parse(s.baseURL + apiPath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %w", err)
	}

	// Add query parameters
	q := u.Query()
	q.Set("page", strconv.Itoa(vietlott.DefaultPageNumber))
	q.Set("pageSize", strconv.Itoa(limit))
	u.RawQuery = q.Encode()

	// Make request with retry
	var resp *http.Response
	for attempt := 0; attempt < s.retryCount; attempt++ {
		req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		// Set headers
		req.Header.Set("Accept", "application/json")
		req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; tool_predict/1.0)")

		resp, err = s.client.Do(req)
		if err == nil && resp.StatusCode == http.StatusOK {
			break
		}

		if resp != nil {
			resp.Body.Close()
		}

		if attempt < s.retryCount-1 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(time.Second * time.Duration(attempt+1)):
				// Exponential backoff
			}
		}
	}

	if err != nil {
		return nil, fmt.Errorf("failed to fetch from API after %d attempts: %w", s.retryCount, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	// Parse response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Try to parse as JSON
	var apiResponse struct {
		Data struct {
			Items []struct {
				DrawNumber int     `json:"drawNumber"`
				Numbers    []int   `json:"numbers"`
				DrawDate   string  `json:"drawDate"`
				Jackpot    float64 `json:"jackpot"`
				Winners    int     `json:"winners"`
			} `json:"items"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &apiResponse); err != nil {
		// Not a valid API response, fall back to web scraping
		return nil, fmt.Errorf("invalid API response: %w", err)
	}

	// Convert to entities
	draws := make([]*entity.Draw, 0, len(apiResponse.Data.Items))
	for _, item := range apiResponse.Data.Items {
		numbers, err := valueobject.NewNumbers(item.Numbers)
		if err != nil {
			logger.Warn("Invalid numbers in draw",
				zap.Int("draw_number", item.DrawNumber),
				zap.Error(err),
			)
			continue
		}

		drawDate, err := time.Parse("2006-01-02T15:04:05", item.DrawDate)
		if err != nil {
			logger.Warn("Invalid date format in draw",
				zap.Int("draw_number", item.DrawNumber),
				zap.String("date", item.DrawDate),
				zap.Error(err),
			)
			// Try alternative date formats
			drawDate, err = time.Parse("2006-01-02", item.DrawDate)
			if err != nil {
				continue
			}
		}

		draw, err := entity.NewDraw(
			gameType,
			item.DrawNumber,
			numbers,
			drawDate,
			item.Jackpot,
			item.Winners,
		)
		if err != nil {
			logger.Warn("Failed to create draw entity",
				zap.Error(err),
			)
			continue
		}

		draws = append(draws, draw)
	}

	if len(draws) == 0 {
		return nil, fmt.Errorf("no valid draws found in API response")
	}

	return draws, nil
}

// waitForRateLimit implements rate limiting
func (s *VietlottAPIScraper) waitForRateLimit() {
	if s.rateLimit > 0 {
		time.Since(s.lastRequest)
		if time.Since(s.lastRequest) < s.rateLimit {
			time.Sleep(s.rateLimit - time.Since(s.lastRequest))
		}
		s.lastRequest = time.Now()
	}
}

// Ensure VietlottAPIScraper implements port.VietlottScraper
var _ port.VietlottScraper = (*VietlottAPIScraper)(nil)
