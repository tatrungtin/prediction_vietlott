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

	"github.com/tool_predict/internal/domain/entity"
	"github.com/tool_predict/internal/domain/repository"
	"github.com/tool_predict/internal/domain/valueobject"
)

// JSONStorage implements repository.DrawRepository using JSON files
type JSONStorage struct {
	basePath string
	mu       sync.RWMutex
}

// NewJSONStorage creates a new JSON storage adapter
func NewJSONStorage(basePath string) (*JSONStorage, error) {
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create base directory: %w", err)
	}

	// Create subdirectories
	dirs := []string{"draws", "predictions", "ensembles", "backtests", "stats"}
	for _, dir := range dirs {
		if err := os.MkdirAll(filepath.Join(basePath, dir), 0755); err != nil {
			return nil, fmt.Errorf("failed to create %s directory: %w", dir, err)
		}
	}

	return &JSONStorage{
		basePath: basePath,
	}, nil
}

// Save saves a draw to JSON file
func (s *JSONStorage) Save(ctx context.Context, draw *entity.Draw) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	filename := s.getDrawFilename(draw.GameType, draw.ID)
	return s.saveToFile(filename, draw)
}

// SaveBatch saves multiple draws
func (s *JSONStorage) SaveBatch(ctx context.Context, draws []*entity.Draw) error {
	for _, draw := range draws {
		if err := s.Save(ctx, draw); err != nil {
			return err
		}
	}
	return nil
}

// FindByID finds a draw by ID
func (s *JSONStorage) FindByID(ctx context.Context, id string) (*entity.Draw, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Search in all game type directories
	gameTypes := []valueobject.GameType{valueobject.Mega645, valueobject.Power655}
	for _, gameType := range gameTypes {
		filename := s.getDrawFilename(gameType, id)
		if _, err := os.Stat(filename); err == nil {
			var draw entity.Draw
			if err := s.loadFromFile(filename, &draw); err != nil {
				return nil, err
			}
			return &draw, nil
		}
	}

	return nil, fmt.Errorf("draw with ID %s not found", id)
}

// FindByGameTypeAndDrawNumber finds a draw by game type and draw number
func (s *JSONStorage) FindByGameTypeAndDrawNumber(
	ctx context.Context,
	gameType valueobject.GameType,
	drawNumber int,
) (*entity.Draw, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	dir := s.getGameTypeDir("draws", gameType)
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		var draw entity.Draw
		filename := filepath.Join(dir, file.Name())
		if err := s.loadFromFile(filename, &draw); err != nil {
			continue
		}

		if draw.DrawNumber == drawNumber {
			return &draw, nil
		}
	}

	return nil, fmt.Errorf("draw number %d not found for game type %s", drawNumber, gameType)
}

// FindLatest finds the most recent draws
func (s *JSONStorage) FindLatest(
	ctx context.Context,
	gameType valueobject.GameType,
	limit int,
) ([]*entity.Draw, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	dir := s.getGameTypeDir("draws", gameType)
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	draws := make([]*entity.Draw, 0, limit)
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		var draw entity.Draw
		filename := filepath.Join(dir, file.Name())
		if err := s.loadFromFile(filename, &draw); err != nil {
			continue
		}

		draws = append(draws, &draw)
	}

	// Sort by draw date (descending) and limit
	sortDrawsByDate(draws, false)
	if len(draws) > limit {
		draws = draws[:limit]
	}

	return draws, nil
}

// FindByDateRange finds draws within a date range
func (s *JSONStorage) FindByDateRange(
	ctx context.Context,
	gameType valueobject.GameType,
	dateRange valueobject.DateRange,
) ([]*entity.Draw, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	dir := s.getGameTypeDir("draws", gameType)
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	draws := make([]*entity.Draw, 0)
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		var draw entity.Draw
		filename := filepath.Join(dir, file.Name())
		if err := s.loadFromFile(filename, &draw); err != nil {
			continue
		}

		if dateRange.Contains(draw.DrawDate) {
			draws = append(draws, &draw)
		}
	}

	return draws, nil
}

// Count returns the total number of draws for a game type
func (s *JSONStorage) Count(ctx context.Context, gameType valueobject.GameType) (int64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	dir := s.getGameTypeDir("draws", gameType)
	files, err := os.ReadDir(dir)
	if err != nil {
		return 0, err
	}

	count := int64(0)
	for _, file := range files {
		if !file.IsDir() {
			count++
		}
	}

	return count, nil
}

// DeleteAll deletes all draws for a game type
func (s *JSONStorage) DeleteAll(ctx context.Context, gameType valueobject.GameType) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	dir := s.getGameTypeDir("draws", gameType)
	files, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	for _, file := range files {
		if !file.IsDir() {
			filename := filepath.Join(dir, file.Name())
			if err := os.Remove(filename); err != nil {
				return err
			}
		}
	}

	return nil
}

// GetLatestDrawNumber returns the highest draw number
func (s *JSONStorage) GetLatestDrawNumber(ctx context.Context, gameType valueobject.GameType) (int, error) {
	draws, err := s.FindLatest(ctx, gameType, 1)
	if err != nil {
		return 0, err
	}

	if len(draws) == 0 {
		return 0, fmt.Errorf("no draws found for game type %s", gameType)
	}

	return draws[0].DrawNumber, nil
}

// FindByDrawNumberRange finds draws within a draw number range
func (s *JSONStorage) FindByDrawNumberRange(
	ctx context.Context,
	gameType valueobject.GameType,
	startDrawNumber int,
	endDrawNumber int,
) ([]*entity.Draw, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	dir := s.getGameTypeDir("draws", gameType)
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	draws := make([]*entity.Draw, 0)
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		var draw entity.Draw
		filename := filepath.Join(dir, file.Name())
		if err := s.loadFromFile(filename, &draw); err != nil {
			continue
		}

		if draw.DrawNumber >= startDrawNumber && draw.DrawNumber <= endDrawNumber {
			draws = append(draws, &draw)
		}
	}

	return draws, nil
}

// Helper methods

func (s *JSONStorage) getDrawFilename(gameType valueobject.GameType, id string) string {
	return filepath.Join(s.getGameTypeDir("draws", gameType), id+".json")
}

func (s *JSONStorage) getGameTypeDir(subDir string, gameType valueobject.GameType) string {
	gameTypeStr := strings.ToLower(string(gameType))
	return filepath.Join(s.basePath, subDir, gameTypeStr)
}

func (s *JSONStorage) saveToFile(filename string, data interface{}) error {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filename, jsonData, 0644)
}

func (s *JSONStorage) loadFromFile(filename string, data interface{}) error {
	file, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	return json.Unmarshal(file, data)
}

func sortDrawsByDate(draws []*entity.Draw, ascending bool) {
	sort.Slice(draws, func(i, j int) bool {
		if ascending {
			return draws[i].DrawDate.Before(draws[j].DrawDate)
		}
		return draws[i].DrawDate.After(draws[j].DrawDate)
	})
}

// Ensure JSONStorage implements repository.DrawRepository
var _ repository.DrawRepository = (*JSONStorage)(nil)
