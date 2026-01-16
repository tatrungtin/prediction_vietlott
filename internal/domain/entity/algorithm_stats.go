package entity

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/tool_predict/internal/domain/valueobject"
)

// AlgorithmStats represents performance statistics for a prediction algorithm
type AlgorithmStats struct {
	ID              string                   `json:"id"`
	AlgorithmName   string                   `json:"algorithm_name"`
	GameType        valueobject.GameType     `json:"game_type"`
	TotalPredictions int                     `json:"total_predictions"`

	// Performance metrics
	Accuracy3Numbers float64                  `json:"accuracy_3_numbers"`
	Accuracy4Numbers float64                  `json:"accuracy_4_numbers"`
	AccuracyExact    float64                  `json:"accuracy_exact"`
	AverageConfidence float64                 `json:"average_confidence"`

	// Metadata
	IsActive        bool                      `json:"is_active"`
	Weight          float64                  `json:"weight"` // For ensemble voting
	LastUpdated     time.Time                `json:"last_updated"`
	CreatedAt       time.Time                `json:"created_at"`
}

// NewAlgorithmStats creates a new AlgorithmStats entity
func NewAlgorithmStats(
	algorithmName string,
	gameType valueobject.GameType,
	weight float64,
) (*AlgorithmStats, error) {
	// Validate algorithm name
	if algorithmName == "" {
		return nil, fmt.Errorf("algorithm name cannot be empty")
	}

	// Validate game type
	if err := gameType.Validate(); err != nil {
		return nil, fmt.Errorf("invalid game type: %w", err)
	}

	// Validate weight
	if weight < 0 {
		return nil, fmt.Errorf("weight cannot be negative, got %f", weight)
	}

	now := time.Now()
	return &AlgorithmStats{
		ID:               uuid.New().String(),
		AlgorithmName:    algorithmName,
		GameType:         gameType,
		TotalPredictions: 0,
		Accuracy3Numbers: 0.0,
		Accuracy4Numbers: 0.0,
		AccuracyExact:    0.0,
		AverageConfidence: 0.0,
		IsActive:         true,
		Weight:           weight,
		LastUpdated:      now,
		CreatedAt:        now,
	}, nil
}

// UpdateMetrics updates the performance metrics based on backtest results
func (as *AlgorithmStats) UpdateMetrics(
	accuracy3Num float64,
	accuracy4Num float64,
	accuracyExact float64,
	avgConfidence float64,
	totalPredictions int,
) {
	as.Accuracy3Numbers = accuracy3Num
	as.Accuracy4Numbers = accuracy4Num
	as.AccuracyExact = accuracyExact
	as.AverageConfidence = avgConfidence
	as.TotalPredictions = totalPredictions
	as.LastUpdated = time.Now()
}

// SetWeight updates the algorithm's weight for ensemble voting
func (as *AlgorithmStats) SetWeight(weight float64) error {
	if weight < 0 {
		return fmt.Errorf("weight cannot be negative, got %f", weight)
	}
	as.Weight = weight
	as.LastUpdated = time.Now()
	return nil
}

// SetActive sets the active status of the algorithm
func (as *AlgorithmStats) SetActive(isActive bool) {
	as.IsActive = isActive
	as.LastUpdated = time.Now()
}

// GetOverallScore returns a weighted overall score for the algorithm
func (as *AlgorithmStats) GetOverallScore() float64 {
	// Simple weighted average: exact matches worth most, then 4-number, then 3-number
	score := (as.AccuracyExact * 0.5) +
		(as.Accuracy4Numbers * 0.3) +
		(as.Accuracy3Numbers * 0.2)
	return score
}

// String returns a string representation of the algorithm stats
func (as *AlgorithmStats) String() string {
	return fmt.Sprintf("AlgorithmStats: %s (%s) - Weight: %.2f, Exact Acc: %.2f%%, 4-Num Acc: %.2f%%, 3-Num Acc: %.2f%%, Active: %v",
		as.AlgorithmName,
		as.GameType,
		as.Weight,
		as.AccuracyExact*100,
		as.Accuracy4Numbers*100,
		as.Accuracy3Numbers*100,
		as.IsActive,
	)
}
