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

// PredictionJSONStorage implements repository.PredictionRepository
type PredictionJSONStorage struct {
	basePath string
	mu       sync.RWMutex
}

// NewPredictionJSONStorage creates a new prediction storage adapter
func NewPredictionJSONStorage(basePath string) (*PredictionJSONStorage, error) {
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create base directory: %w", err)
	}

	// Create subdirectories
	dirs := []string{"predictions", "ensembles"}
	for _, dir := range dirs {
		if err := os.MkdirAll(filepath.Join(basePath, dir), 0755); err != nil {
			return nil, fmt.Errorf("failed to create %s directory: %w", dir, err)
		}
	}

	return &PredictionJSONStorage{
		basePath: basePath,
	}, nil
}

// Save saves a single prediction
func (s *PredictionJSONStorage) Save(ctx context.Context, prediction *entity.Prediction) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	filename := s.getPredictionFilename(prediction.GameType, prediction.ID)
	return s.saveToFile(filename, prediction)
}

// SaveBatch saves multiple predictions
func (s *PredictionJSONStorage) SaveBatch(ctx context.Context, predictions []*entity.Prediction) error {
	for _, pred := range predictions {
		if err := s.Save(ctx, pred); err != nil {
			return err
		}
	}
	return nil
}

// SaveEnsemble saves an ensemble prediction
func (s *PredictionJSONStorage) SaveEnsemble(ctx context.Context, ensemble *entity.EnsemblePrediction) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	filename := s.getEnsembleFilename(ensemble.GameType, ensemble.ID)
	return s.saveToFile(filename, ensemble)
}

// FindByID finds a prediction by ID
func (s *PredictionJSONStorage) FindByID(ctx context.Context, id string) (*entity.Prediction, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Search in all game type directories
	gameTypes := []valueobject.GameType{valueobject.Mega645, valueobject.Power655}
	for _, gameType := range gameTypes {
		dir := s.getGameTypeDir("predictions", gameType)
		files, err := os.ReadDir(dir)
		if err != nil {
			continue
		}

		for _, file := range files {
			if file.IsDir() {
				continue
			}

			var pred entity.Prediction
			filename := filepath.Join(dir, file.Name())
			if err := s.loadFromFile(filename, &pred); err != nil {
				continue
			}

			if pred.ID == id {
				return &pred, nil
			}
		}
	}

	return nil, fmt.Errorf("prediction with ID %s not found", id)
}

// FindEnsembleByID finds an ensemble prediction by ID
func (s *PredictionJSONStorage) FindEnsembleByID(ctx context.Context, id string) (*entity.EnsemblePrediction, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Search in all game type directories
	gameTypes := []valueobject.GameType{valueobject.Mega645, valueobject.Power655}
	for _, gameType := range gameTypes {
		dir := s.getGameTypeDir("ensembles", gameType)
		files, err := os.ReadDir(dir)
		if err != nil {
			continue
		}

		for _, file := range files {
			if file.IsDir() {
				continue
			}

			var ensemble entity.EnsemblePrediction
			filename := filepath.Join(dir, file.Name())
			if err := s.loadFromFile(filename, &ensemble); err != nil {
				continue
			}

			if ensemble.ID == id {
				return &ensemble, nil
			}
		}
	}

	return nil, fmt.Errorf("ensemble prediction with ID %s not found", id)
}

// FindLatest finds the most recent predictions for a game type
func (s *PredictionJSONStorage) FindLatest(
	ctx context.Context,
	gameType valueobject.GameType,
	limit int,
) ([]*entity.Prediction, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	dir := s.getGameTypeDir("predictions", gameType)
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	predictions := make([]*entity.Prediction, 0, limit)
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		var pred entity.Prediction
		filename := filepath.Join(dir, file.Name())
		if err := s.loadFromFile(filename, &pred); err != nil {
			continue
		}

		predictions = append(predictions, &pred)
	}

	// Sort by generated date (descending) and limit
	sortPredictionsByDate(predictions, false)
	if len(predictions) > limit {
		predictions = predictions[:limit]
	}

	return predictions, nil
}

// FindLatestEnsembles finds the most recent ensemble predictions for a game type
func (s *PredictionJSONStorage) FindLatestEnsembles(
	ctx context.Context,
	gameType valueobject.GameType,
	limit int,
) ([]*entity.EnsemblePrediction, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	dir := s.getGameTypeDir("ensembles", gameType)
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	ensembles := make([]*entity.EnsemblePrediction, 0, limit)
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		var ensemble entity.EnsemblePrediction
		filename := filepath.Join(dir, file.Name())
		if err := s.loadFromFile(filename, &ensemble); err != nil {
			continue
		}

		ensembles = append(ensembles, &ensemble)
	}

	// Sort by generated date (descending) and limit
	sortEnsemblesByDate(ensembles, false)
	if len(ensembles) > limit {
		ensembles = ensembles[:limit]
	}

	return ensembles, nil
}

// FindByAlgorithm finds predictions for a specific algorithm and game type
func (s *PredictionJSONStorage) FindByAlgorithm(
	ctx context.Context,
	algorithmName string,
	gameType valueobject.GameType,
	limit int,
) ([]*entity.Prediction, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	dir := s.getGameTypeDir("predictions", gameType)
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	predictions := make([]*entity.Prediction, 0, limit)
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		var pred entity.Prediction
		filename := filepath.Join(dir, file.Name())
		if err := s.loadFromFile(filename, &pred); err != nil {
			continue
		}

		if pred.AlgorithmName == algorithmName {
			predictions = append(predictions, &pred)
		}
	}

	// Sort by generated date (descending) and limit
	sortPredictionsByDate(predictions, false)
	if len(predictions) > limit {
		predictions = predictions[:limit]
	}

	return predictions, nil
}

// FindByDateRange finds predictions within a date range for a game type
func (s *PredictionJSONStorage) FindByDateRange(
	ctx context.Context,
	gameType valueobject.GameType,
	startDate interface{},
	endDate interface{},
) ([]*entity.Prediction, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	start, ok1 := startDate.(time.Time)
	end, ok2 := endDate.(time.Time)
	if !ok1 || !ok2 {
		return nil, fmt.Errorf("startDate and endDate must be time.Time")
	}

	dir := s.getGameTypeDir("predictions", gameType)
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	predictions := make([]*entity.Prediction, 0)
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		var pred entity.Prediction
		filename := filepath.Join(dir, file.Name())
		if err := s.loadFromFile(filename, &pred); err != nil {
			continue
		}

		if (pred.GeneratedAt.Equal(start) || pred.GeneratedAt.After(start)) &&
			pred.GeneratedAt.Before(end) {
			predictions = append(predictions, &pred)
		}
	}

	return predictions, nil
}

// Count returns the total number of predictions for a game type
func (s *PredictionJSONStorage) Count(ctx context.Context, gameType valueobject.GameType) (int64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	dir := s.getGameTypeDir("predictions", gameType)
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

// DeleteOld removes predictions older than a certain date
func (s *PredictionJSONStorage) DeleteOld(ctx context.Context, beforeDate interface{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	before, ok := beforeDate.(time.Time)
	if !ok {
		return fmt.Errorf("beforeDate must be time.Time")
	}

	// Delete from both game types
	gameTypes := []valueobject.GameType{valueobject.Mega645, valueobject.Power655}
	for _, gameType := range gameTypes {
		dir := s.getGameTypeDir("predictions", gameType)
		files, err := os.ReadDir(dir)
		if err != nil {
			continue
		}

		for _, file := range files {
			if file.IsDir() {
				continue
			}

			var pred entity.Prediction
			filename := filepath.Join(dir, file.Name())
			if err := s.loadFromFile(filename, &pred); err != nil {
				continue
			}

			if pred.GeneratedAt.Before(before) {
				os.Remove(filename)
			}
		}
	}

	return nil
}

// Helper methods

func (s *PredictionJSONStorage) getPredictionFilename(gameType valueobject.GameType, id string) string {
	return filepath.Join(s.getGameTypeDir("predictions", gameType), id+".json")
}

func (s *PredictionJSONStorage) getEnsembleFilename(gameType valueobject.GameType, id string) string {
	return filepath.Join(s.getGameTypeDir("ensembles", gameType), id+".json")
}

func (s *PredictionJSONStorage) getGameTypeDir(subDir string, gameType valueobject.GameType) string {
	gameTypeStr := strings.ToLower(string(gameType))
	return filepath.Join(s.basePath, subDir, gameTypeStr)
}

func (s *PredictionJSONStorage) saveToFile(filename string, data interface{}) error {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filename, jsonData, 0644)
}

func (s *PredictionJSONStorage) loadFromFile(filename string, data interface{}) error {
	file, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	return json.Unmarshal(file, data)
}

func sortPredictionsByDate(predictions []*entity.Prediction, ascending bool) {
	sort.Slice(predictions, func(i, j int) bool {
		if ascending {
			return predictions[i].GeneratedAt.Before(predictions[j].GeneratedAt)
		}
		return predictions[i].GeneratedAt.After(predictions[j].GeneratedAt)
	})
}

func sortEnsemblesByDate(ensembles []*entity.EnsemblePrediction, ascending bool) {
	sort.Slice(ensembles, func(i, j int) bool {
		if ascending {
			return ensembles[i].GeneratedAt.Before(ensembles[j].GeneratedAt)
		}
		return ensembles[i].GeneratedAt.After(ensembles[j].GeneratedAt)
	})
}

// Ensure PredictionJSONStorage implements repository.PredictionRepository
var _ repository.PredictionRepository = (*PredictionJSONStorage)(nil)
