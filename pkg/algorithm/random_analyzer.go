package algorithm

import (
	"context"
	"fmt"
	"math/rand/v2"
	"sync"
	"time"

	"github.com/tool_predict/internal/domain/entity"
	"github.com/tool_predict/internal/domain/valueobject"
)

// RandomAnalyzer generates purely random predictions
type RandomAnalyzer struct {
	name     string
	weight   float64
	minDraws int
	mu       sync.RWMutex
}

// NewRandomAnalyzer creates a new random analyzer
func NewRandomAnalyzer(weight float64) *RandomAnalyzer {
	return &RandomAnalyzer{
		name:     "random_analysis",
		weight:   weight,
		minDraws: 0, // No minimum draws required for random
	}
}

// Name returns the algorithm name
func (ra *RandomAnalyzer) Name() string {
	return ra.name
}

// GetWeight returns the algorithm's weight
func (ra *RandomAnalyzer) GetWeight() float64 {
	ra.mu.RLock()
	defer ra.mu.RUnlock()
	return ra.weight
}

// SetWeight sets the algorithm's weight
func (ra *RandomAnalyzer) SetWeight(weight float64) error {
	if weight < 0 {
		return fmt.Errorf("weight cannot be negative, got %f", weight)
	}
	ra.mu.Lock()
	defer ra.mu.Unlock()
	ra.weight = weight
	return nil
}

// Validate checks if there's enough data for prediction
func (ra *RandomAnalyzer) Validate(historicalData []*entity.Draw) error {
	// Random analysis doesn't require any historical data
	return nil
}

// Train updates algorithm parameters (random doesn't need training)
func (ra *RandomAnalyzer) Train(ctx context.Context, historicalData []*entity.Draw) error {
	return nil
}

// Predict generates purely random predictions
func (ra *RandomAnalyzer) Predict(
	ctx context.Context,
	gameType valueobject.GameType,
	historicalData []*entity.Draw,
) (*entity.Prediction, error) {
	// Check for cancellation
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// Get number range for game type
	minRange, maxRange := gameType.NumberRange()

	// Generate unique random numbers
	predictedNums := make([]int, 0, 6)
	used := make(map[int]bool)

	for len(predictedNums) < 6 {
		// Generate random number in range [minRange, maxRange]
		num := rand.IntN(maxRange-minRange+1) + minRange

		if !used[num] {
			used[num] = true
			predictedNums = append(predictedNums, num)
		}
	}

	// Sort for consistency
	sortNumbers(predictedNums)

	// Create numbers value object
	numbers, err := valueobject.NewNumbers(predictedNums)
	if err != nil {
		return nil, fmt.Errorf("failed to create numbers: %w", err)
	}

	// Confidence is low for random predictions
	confidence := 0.1

	// Create prediction
	prediction := &entity.Prediction{
		ID:            "", // Will be set by repository
		GameType:      gameType,
		AlgorithmName: ra.name,
		Numbers:       numbers,
		Confidence:    confidence,
		GeneratedAt:   time.Now(),
		ForDate:       time.Now().Add(24 * time.Hour), // Predict for tomorrow
		Metadata: map[string]string{
			"min_draws_required": "0",
			"total_draws_used":   fmt.Sprintf("%d", len(historicalData)),
			"type":               "random",
		},
	}

	return prediction, nil
}

// SetMinDraws sets the minimum number of draws required for prediction
func (ra *RandomAnalyzer) SetMinDraws(minDraws int) error {
	// Random analyzer doesn't need minimum draws, but we can set it to 0
	ra.mu.Lock()
	defer ra.mu.Unlock()
	ra.minDraws = minDraws
	return nil
}

// GetMinDraws returns the minimum number of draws required
func (ra *RandomAnalyzer) GetMinDraws() int {
	ra.mu.RLock()
	defer ra.mu.RUnlock()
	return ra.minDraws
}

// sortNumbers sorts a slice of integers in place
func sortNumbers(nums []int) {
	for i := 0; i < len(nums)-1; i++ {
		for j := i + 1; j < len(nums); j++ {
			if nums[i] > nums[j] {
				nums[i], nums[j] = nums[j], nums[i]
			}
		}
	}
}
