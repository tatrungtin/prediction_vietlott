package entity

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/tool_predict/internal/domain/valueobject"
)

// Prediction represents a single algorithm's lottery number prediction
type Prediction struct {
	ID            string                   `json:"id"`
	GameType      valueobject.GameType     `json:"game_type"`
	AlgorithmName string                   `json:"algorithm_name"`
	Numbers       valueobject.Numbers      `json:"numbers"`
	Confidence    float64                  `json:"confidence"`
	GeneratedAt   time.Time                `json:"generated_at"`
	ForDate       time.Time                `json:"for_date"`
	Metadata      map[string]string        `json:"metadata,omitempty"`
}

// NewPrediction creates a new Prediction entity
func NewPrediction(
	gameType valueobject.GameType,
	algorithmName string,
	numbers valueobject.Numbers,
	confidence float64,
	forDate time.Time,
) (*Prediction, error) {
	// Validate game type
	if err := gameType.Validate(); err != nil {
		return nil, fmt.Errorf("invalid game type: %w", err)
	}

	// Validate algorithm name
	if algorithmName == "" {
		return nil, fmt.Errorf("algorithm name cannot be empty")
	}

	// Validate confidence
	if confidence < 0 || confidence > 1 {
		return nil, fmt.Errorf("confidence must be between 0 and 1, got %f", confidence)
	}

	return &Prediction{
		ID:            uuid.New().String(),
		GameType:      gameType,
		AlgorithmName: algorithmName,
		Numbers:       numbers,
		Confidence:    confidence,
		GeneratedAt:   time.Now(),
		ForDate:       forDate,
		Metadata:      make(map[string]string),
	}, nil
}

// AlgorithmContribution represents an algorithm's contribution to ensemble prediction
type AlgorithmContribution struct {
	AlgorithmName string  `json:"algorithm_name"`
	Weight        float64 `json:"weight"`
	MatchCount    int     `json:"match_count"`
	Confidence    float64 `json:"confidence"`
}

// EnsemblePrediction represents a combined prediction from multiple algorithms
type EnsemblePrediction struct {
	ID              string                       `json:"id"`
	GameType        valueobject.GameType         `json:"game_type"`
	Predictions     []*Prediction                `json:"predictions"`
	FinalNumbers    valueobject.Numbers          `json:"final_numbers"`
	VotingStrategy  string                       `json:"voting_strategy"`
	GeneratedAt     time.Time                    `json:"generated_at"`
	AlgorithmStats  []AlgorithmContribution      `json:"algorithm_stats"`
}

// NewEnsemblePrediction creates a new EnsemblePrediction entity
func NewEnsemblePrediction(
	gameType valueobject.GameType,
	predictions []*Prediction,
	finalNumbers valueobject.Numbers,
	votingStrategy string,
	algorithmStats []AlgorithmContribution,
) (*EnsemblePrediction, error) {
	// Validate game type
	if err := gameType.Validate(); err != nil {
		return nil, fmt.Errorf("invalid game type: %w", err)
	}

	// Validate predictions
	if len(predictions) == 0 {
		return nil, fmt.Errorf("at least one prediction is required")
	}

	// Validate voting strategy
	if votingStrategy == "" {
		return nil, fmt.Errorf("voting strategy cannot be empty")
	}

	return &EnsemblePrediction{
		ID:             uuid.New().String(),
		GameType:       gameType,
		Predictions:    predictions,
		FinalNumbers:   finalNumbers,
		VotingStrategy: votingStrategy,
		GeneratedAt:    time.Now(),
		AlgorithmStats: algorithmStats,
	}, nil
}

// GetID returns the unique identifier of the ensemble prediction
func (ep *EnsemblePrediction) GetID() string {
	return ep.ID
}

// GetFinalNumbers returns the final predicted numbers
func (ep *EnsemblePrediction) GetFinalNumbers() valueobject.Numbers {
	return ep.FinalNumbers
}

// String returns a string representation of the ensemble prediction
func (ep *EnsemblePrediction) String() string {
	return fmt.Sprintf("EnsemblePrediction #%s (%s) on %s: %s (strategy: %s, algorithms: %d)",
		ep.ID,
		ep.GameType,
		ep.GeneratedAt.Format("2006-01-02 15:04:05"),
		ep.FinalNumbers,
		ep.VotingStrategy,
		len(ep.Predictions),
	)
}
