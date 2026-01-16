package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/tool_predict/internal/domain/entity"
	"github.com/tool_predict/internal/domain/repository"
	"github.com/tool_predict/internal/domain/valueobject"
)

// StatsJSONStorage implements repository.StatsRepository
type StatsJSONStorage struct {
	basePath string
	mu       sync.RWMutex
}

// NewStatsJSONStorage creates a new stats storage adapter
func NewStatsJSONStorage(basePath string) (*StatsJSONStorage, error) {
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create base directory: %w", err)
	}

	// Create subdirectory
	if err := os.MkdirAll(filepath.Join(basePath, "stats"), 0755); err != nil {
		return nil, fmt.Errorf("failed to create stats directory: %w", err)
	}

	return &StatsJSONStorage{
		basePath: basePath,
	}, nil
}

// Save saves algorithm statistics
func (s *StatsJSONStorage) Save(ctx context.Context, stats *entity.AlgorithmStats) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	filename := s.getStatsFilename(stats.GameType, stats.AlgorithmName)
	return s.saveToFile(filename, stats)
}

// Find finds statistics for a specific algorithm and game type
func (s *StatsJSONStorage) Find(
	ctx context.Context,
	algorithmName string,
	gameType valueobject.GameType,
) (*entity.AlgorithmStats, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	filename := s.getStatsFilename(gameType, algorithmName)
	if _, err := os.Stat(filename); err != nil {
		return nil, fmt.Errorf("stats not found for algorithm %s and game type %s", algorithmName, gameType)
	}

	var stats entity.AlgorithmStats
	if err := s.loadFromFile(filename, &stats); err != nil {
		return nil, err
	}

	return &stats, nil
}

// FindAll finds all algorithm statistics
func (s *StatsJSONStorage) FindAll(ctx context.Context) ([]*entity.AlgorithmStats, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	allStats := make([]*entity.AlgorithmStats, 0)
	gameTypes := []valueobject.GameType{valueobject.Mega645, valueobject.Power655}
	for _, gameType := range gameTypes {
		dir := s.getGameTypeDir("stats", gameType)
		files, err := os.ReadDir(dir)
		if err != nil {
			continue
		}

		for _, file := range files {
			if file.IsDir() {
				continue
			}

			var stats entity.AlgorithmStats
			filename := filepath.Join(dir, file.Name())
			if err := s.loadFromFile(filename, &stats); err != nil {
				continue
			}

			allStats = append(allStats, &stats)
		}
	}

	return allStats, nil
}

// FindByGameType finds all statistics for a specific game type
func (s *StatsJSONStorage) FindByGameType(
	ctx context.Context,
	gameType valueobject.GameType,
) ([]*entity.AlgorithmStats, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	dir := s.getGameTypeDir("stats", gameType)
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	statsList := make([]*entity.AlgorithmStats, 0)
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		var stats entity.AlgorithmStats
		filename := filepath.Join(dir, file.Name())
		if err := s.loadFromFile(filename, &stats); err != nil {
			continue
		}

		statsList = append(statsList, &stats)
	}

	return statsList, nil
}

// FindActive finds all active algorithm statistics
func (s *StatsJSONStorage) FindActive(ctx context.Context) ([]*entity.AlgorithmStats, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	activeStats := make([]*entity.AlgorithmStats, 0)
	gameTypes := []valueobject.GameType{valueobject.Mega645, valueobject.Power655}
	for _, gameType := range gameTypes {
		dir := s.getGameTypeDir("stats", gameType)
		files, err := os.ReadDir(dir)
		if err != nil {
			continue
		}

		for _, file := range files {
			if file.IsDir() {
				continue
			}

			var stats entity.AlgorithmStats
			filename := filepath.Join(dir, file.Name())
			if err := s.loadFromFile(filename, &stats); err != nil {
				continue
			}

			if stats.IsActive {
				activeStats = append(activeStats, &stats)
			}
		}
	}

	return activeStats, nil
}

// Update updates algorithm statistics in the repository
func (s *StatsJSONStorage) Update(ctx context.Context, stats *entity.AlgorithmStats) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	filename := s.getStatsFilename(stats.GameType, stats.AlgorithmName)
	return s.saveToFile(filename, stats)
}

// UpdateWeight updates only the weight for an algorithm
func (s *StatsJSONStorage) UpdateWeight(
	ctx context.Context,
	algorithmName string,
	gameType valueobject.GameType,
	weight float64,
) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	stats, err := s.Find(ctx, algorithmName, gameType)
	if err != nil {
		return err
	}

	if err := stats.SetWeight(weight); err != nil {
		return err
	}

	filename := s.getStatsFilename(gameType, algorithmName)
	return s.saveToFile(filename, stats)
}

// UpdateMetrics updates the performance metrics for an algorithm
func (s *StatsJSONStorage) UpdateMetrics(
	ctx context.Context,
	algorithmName string,
	gameType valueobject.GameType,
	accuracy3Num float64,
	accuracy4Num float64,
	accuracyExact float64,
	avgConfidence float64,
	totalPredictions int,
) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	stats, err := s.Find(ctx, algorithmName, gameType)
	if err != nil {
		return err
	}

	stats.UpdateMetrics(accuracy3Num, accuracy4Num, accuracyExact, avgConfidence, totalPredictions)

	filename := s.getStatsFilename(gameType, algorithmName)
	return s.saveToFile(filename, stats)
}

// SetActive sets the active status for an algorithm
func (s *StatsJSONStorage) SetActive(
	ctx context.Context,
	algorithmName string,
	gameType valueobject.GameType,
	isActive bool,
) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	stats, err := s.Find(ctx, algorithmName, gameType)
	if err != nil {
		return err
	}

	stats.SetActive(isActive)

	filename := s.getStatsFilename(gameType, algorithmName)
	return s.saveToFile(filename, stats)
}

// Helper methods

func (s *StatsJSONStorage) getStatsFilename(gameType valueobject.GameType, algorithmName string) string {
	// Use algorithm name as filename (sanitize it)
	safeName := strings.ReplaceAll(algorithmName, " ", "_")
	return filepath.Join(s.getGameTypeDir("stats", gameType), safeName+".json")
}

func (s *StatsJSONStorage) getGameTypeDir(subDir string, gameType valueobject.GameType) string {
	gameTypeStr := strings.ToLower(string(gameType))
	return filepath.Join(s.basePath, subDir, gameTypeStr)
}

func (s *StatsJSONStorage) saveToFile(filename string, data interface{}) error {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filename, jsonData, 0644)
}

func (s *StatsJSONStorage) loadFromFile(filename string, data interface{}) error {
	file, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	return json.Unmarshal(file, data)
}

// Ensure StatsJSONStorage implements repository.StatsRepository
var _ repository.StatsRepository = (*StatsJSONStorage)(nil)
