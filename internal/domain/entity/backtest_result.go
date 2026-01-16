package entity

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/tool_predict/internal/domain/valueobject"
)

// PredictionMatch represents a single prediction vs actual result comparison
type PredictionMatch struct {
	PredictedNumbers valueobject.Numbers `json:"predicted_numbers"`
	ActualNumbers    valueobject.Numbers `json:"actual_numbers"`
	MatchCount       int                 `json:"match_count"`
	Confidence       float64             `json:"confidence"`
	PredictionDate   time.Time           `json:"prediction_date"`
	ActualDrawDate   time.Time           `json:"actual_draw_date"`
}

// BacktestResult represents the results of backtesting an algorithm
type BacktestResult struct {
	ID               string                `json:"id"`
	GameType         valueobject.GameType  `json:"game_type"`
	AlgorithmName    string                `json:"algorithm_name"`
	TestPeriod       valueobject.DateRange `json:"test_period"`
	TotalPredictions int                   `json:"total_predictions"`

	// Match statistics
	ExactMatches       int `json:"exact_matches"`
	ThreeNumberMatches int `json:"three_number_matches"`
	FourNumberMatches  int `json:"four_number_matches"`

	// Performance metrics
	AverageConfidence float64       `json:"average_confidence"`
	ExecutionTime     time.Duration `json:"execution_time"`
	CreatedAt         time.Time     `json:"created_at"`
	LastUpdated       time.Time     `json:"last_updated"`

	// Detailed results (optional, can be large)
	DetailedResults []PredictionMatch `json:"detailed_results,omitempty"`
}

// NewBacktestResult creates a new BacktestResult entity
func NewBacktestResult(
	gameType valueobject.GameType,
	algorithmName string,
	testPeriod valueobject.DateRange,
	totalPredictions int,
) (*BacktestResult, error) {
	// Validate game type
	if err := gameType.Validate(); err != nil {
		return nil, fmt.Errorf("invalid game type: %w", err)
	}

	// Validate algorithm name
	if algorithmName == "" {
		return nil, fmt.Errorf("algorithm name cannot be empty")
	}

	// Validate total predictions
	if totalPredictions < 0 {
		return nil, fmt.Errorf("total predictions cannot be negative, got %d", totalPredictions)
	}

	now := time.Now()
	return &BacktestResult{
		ID:                 uuid.New().String(),
		GameType:           gameType,
		AlgorithmName:      algorithmName,
		TestPeriod:         testPeriod,
		TotalPredictions:   totalPredictions,
		ExactMatches:       0,
		ThreeNumberMatches: 0,
		FourNumberMatches:  0,
		AverageConfidence:  0.0,
		ExecutionTime:      0,
		CreatedAt:          now,
		LastUpdated:        now,
		DetailedResults:    make([]PredictionMatch, 0),
	}, nil
}

// AddMatchResult adds a prediction match result to the backtest
func (br *BacktestResult) AddMatchResult(match PredictionMatch) {
	br.DetailedResults = append(br.DetailedResults, match)
	br.LastUpdated = time.Now()

	// Update match counters
	switch match.MatchCount {
	case 3:
		br.ThreeNumberMatches++
	case 4:
		br.FourNumberMatches++
	case 6:
		br.ExactMatches++
	}
}

// CalculateMetrics calculates performance metrics from detailed results
func (br *BacktestResult) CalculateMetrics() {
	if len(br.DetailedResults) == 0 {
		br.AverageConfidence = 0.0
		return
	}

	totalConfidence := 0.0
	for _, result := range br.DetailedResults {
		totalConfidence += result.Confidence
	}
	br.AverageConfidence = totalConfidence / float64(len(br.DetailedResults))
}

// GetAccuracyRate returns the exact match accuracy rate
func (br *BacktestResult) GetAccuracyRate() float64 {
	if br.TotalPredictions == 0 {
		return 0.0
	}
	return float64(br.ExactMatches) / float64(br.TotalPredictions)
}

// GetThreeNumberAccuracy returns the 3+ number match accuracy rate
func (br *BacktestResult) GetThreeNumberAccuracy() float64 {
	if br.TotalPredictions == 0 {
		return 0.0
	}
	return float64(br.ThreeNumberMatches) / float64(br.TotalPredictions)
}

// GetFourNumberAccuracy returns the 4+ number match accuracy rate
func (br *BacktestResult) GetFourNumberAccuracy() float64 {
	if br.TotalPredictions == 0 {
		return 0.0
	}
	return float64(br.FourNumberMatches) / float64(br.TotalPredictions)
}

// String returns a string representation of the backtest result
func (br *BacktestResult) String() string {
	return fmt.Sprintf("BacktestResult #%s: %s - %s, Accuracy: %.2f%% (%d/%d exact matches)",
		br.ID,
		br.AlgorithmName,
		br.GameType,
		br.GetAccuracyRate()*100,
		br.ExactMatches,
		br.TotalPredictions,
	)
}
