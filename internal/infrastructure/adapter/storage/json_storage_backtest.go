package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/tool_predict/internal/domain/entity"
	"github.com/tool_predict/internal/domain/repository"
	"github.com/tool_predict/internal/domain/valueobject"
)

// BacktestJSONStorage implements repository.BacktestRepository
type BacktestJSONStorage struct {
	basePath string
	mu       sync.RWMutex
}

// NewBacktestJSONStorage creates a new backtest storage adapter
func NewBacktestJSONStorage(basePath string) (*BacktestJSONStorage, error) {
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create base directory: %w", err)
	}

	// Create subdirectory
	if err := os.MkdirAll(filepath.Join(basePath, "backtests"), 0755); err != nil {
		return nil, fmt.Errorf("failed to create backtests directory: %w", err)
	}

	return &BacktestJSONStorage{
		basePath: basePath,
	}, nil
}

// Save saves a backtest result
func (s *BacktestJSONStorage) Save(ctx context.Context, result *entity.BacktestResult) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	filename := s.getBacktestFilename(result.GameType, result.ID)
	return s.saveToFile(filename, result)
}

// FindByID finds a backtest result by its unique identifier
func (s *BacktestJSONStorage) FindByID(ctx context.Context, id string) (*entity.BacktestResult, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Search in all game type directories
	gameTypes := []valueobject.GameType{valueobject.Mega645, valueobject.Power655}
	for _, gameType := range gameTypes {
		dir := s.getGameTypeDir("backtests", gameType)
		files, err := os.ReadDir(dir)
		if err != nil {
			continue
		}

		for _, file := range files {
			if file.IsDir() {
				continue
			}

			var result entity.BacktestResult
			filename := filepath.Join(dir, file.Name())
			if err := s.loadFromFile(filename, &result); err != nil {
				continue
			}

			if result.ID == id {
				return &result, nil
			}
		}
	}

	return nil, fmt.Errorf("backtest result with ID %s not found", id)
}

// FindLatest finds the most recent backtest results
func (s *BacktestJSONStorage) FindLatest(ctx context.Context, limit int) ([]*entity.BacktestResult, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	results := make([]*entity.BacktestResult, 0)
	gameTypes := []valueobject.GameType{valueobject.Mega645, valueobject.Power655}
	for _, gameType := range gameTypes {
		dir := s.getGameTypeDir("backtests", gameType)
		files, err := os.ReadDir(dir)
		if err != nil {
			continue
		}

		for _, file := range files {
			if file.IsDir() {
				continue
			}

			var result entity.BacktestResult
			filename := filepath.Join(dir, file.Name())
			if err := s.loadFromFile(filename, &result); err != nil {
				continue
			}

			results = append(results, &result)
		}
	}

	// Sort by test period end date (descending) and limit
	sortBacktestsByDate(results, false)
	if len(results) > limit {
		results = results[:limit]
	}

	return results, nil
}

// FindByAlgorithm finds backtest results for a specific algorithm
func (s *BacktestJSONStorage) FindByAlgorithm(
	ctx context.Context,
	algorithmName string,
	gameType valueobject.GameType,
	limit int,
) ([]*entity.BacktestResult, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	dir := s.getGameTypeDir("backtests", gameType)
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	results := make([]*entity.BacktestResult, 0, limit)
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		var result entity.BacktestResult
		filename := filepath.Join(dir, file.Name())
		if err := s.loadFromFile(filename, &result); err != nil {
			continue
		}

		if result.AlgorithmName == algorithmName {
			results = append(results, &result)
		}
	}

	// Sort by test period end date (descending) and limit
	sortBacktestsByDate(results, false)
	if len(results) > limit {
		results = results[:limit]
	}

	return results, nil
}

// FindByGameType finds all backtest results for a specific game type
func (s *BacktestJSONStorage) FindByGameType(
	ctx context.Context,
	gameType valueobject.GameType,
) ([]*entity.BacktestResult, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	dir := s.getGameTypeDir("backtests", gameType)
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	results := make([]*entity.BacktestResult, 0)
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		var result entity.BacktestResult
		filename := filepath.Join(dir, file.Name())
		if err := s.loadFromFile(filename, &result); err != nil {
			continue
		}

		results = append(results, &result)
	}

	return results, nil
}

// FindByDateRange finds backtest results within a date range
func (s *BacktestJSONStorage) FindByDateRange(
	ctx context.Context,
	startDate interface{},
	endDate interface{},
) ([]*entity.BacktestResult, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	start, ok1 := startDate.(time.Time)
	end, ok2 := endDate.(time.Time)
	if !ok1 || !ok2 {
		return nil, fmt.Errorf("startDate and endDate must be time.Time")
	}

	results := make([]*entity.BacktestResult, 0)
	gameTypes := []valueobject.GameType{valueobject.Mega645, valueobject.Power655}
	for _, gameType := range gameTypes {
		dir := s.getGameTypeDir("backtests", gameType)
		files, err := os.ReadDir(dir)
		if err != nil {
			continue
		}

		for _, file := range files {
			if file.IsDir() {
				continue
			}

			var result entity.BacktestResult
			filename := filepath.Join(dir, file.Name())
			if err := s.loadFromFile(filename, &result); err != nil {
				continue
			}

			// Check if test period overlaps with date range
			if (result.TestPeriod.EndDate.Equal(start) || result.TestPeriod.EndDate.After(start)) &&
				result.TestPeriod.StartDate.Before(end) {
				results = append(results, &result)
			}
		}
	}

	return results, nil
}

// FindBestPerforming finds the best performing algorithm for a game type
func (s *BacktestJSONStorage) FindBestPerforming(
	ctx context.Context,
	gameType valueobject.GameType,
	metric string,
) (*entity.BacktestResult, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	dir := s.getGameTypeDir("backtests", gameType)
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var bestResult *entity.BacktestResult
	var bestScore float64

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		var result entity.BacktestResult
		filename := filepath.Join(dir, file.Name())
		if err := s.loadFromFile(filename, &result); err != nil {
			continue
		}

		// Calculate score based on metric
		var score float64
		switch metric {
		case "exact":
			score = float64(result.ExactMatches)
		case "4_numbers":
			score = float64(result.FourNumberMatches)
		case "3_numbers":
			score = float64(result.ThreeNumberMatches)
		default:
			continue
		}

		if bestResult == nil || score > bestScore {
			bestResult = &result
			bestScore = score
		}
	}

	if bestResult == nil {
		return nil, fmt.Errorf("no backtest results found for game type %s", gameType)
	}

	return bestResult, nil
}

// DeleteOld removes backtest results older than a certain date
func (s *BacktestJSONStorage) DeleteOld(ctx context.Context, beforeDate interface{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	before, ok := beforeDate.(time.Time)
	if !ok {
		return fmt.Errorf("beforeDate must be time.Time")
	}

	// Delete from both game types
	gameTypes := []valueobject.GameType{valueobject.Mega645, valueobject.Power655}
	for _, gameType := range gameTypes {
		dir := s.getGameTypeDir("backtests", gameType)
		files, err := os.ReadDir(dir)
		if err != nil {
			continue
		}

		for _, file := range files {
			if file.IsDir() {
				continue
			}

			var result entity.BacktestResult
			filename := filepath.Join(dir, file.Name())
			if err := s.loadFromFile(filename, &result); err != nil {
				continue
			}

			if result.TestPeriod.EndDate.Before(before) {
				os.Remove(filename)
			}
		}
	}

	return nil
}

// Helper methods

func (s *BacktestJSONStorage) getBacktestFilename(gameType valueobject.GameType, id string) string {
	return filepath.Join(s.getGameTypeDir("backtests", gameType), id+".json")
}

func (s *BacktestJSONStorage) getGameTypeDir(subDir string, gameType valueobject.GameType) string {
	gameTypeStr := strings.ToLower(string(gameType))
	return filepath.Join(s.basePath, subDir, gameTypeStr)
}

func (s *BacktestJSONStorage) saveToFile(filename string, data interface{}) error {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filename, jsonData, 0644)
}

func (s *BacktestJSONStorage) loadFromFile(filename string, data interface{}) error {
	file, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	return json.Unmarshal(file, data)
}

func sortBacktestsByDate(results []*entity.BacktestResult, ascending bool) {
	sort.Slice(results, func(i, j int) bool {
		if ascending {
			return results[i].TestPeriod.EndDate.Before(results[j].TestPeriod.EndDate)
		}
		return results[i].TestPeriod.EndDate.After(results[j].TestPeriod.EndDate)
	})
}

// Ensure BacktestJSONStorage implements repository.BacktestRepository
var _ repository.BacktestRepository = (*BacktestJSONStorage)(nil)
