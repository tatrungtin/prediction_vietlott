package algorithm

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/tool_predict/internal/domain/entity"
	"github.com/tool_predict/internal/domain/valueobject"
)

// FrequencyAnalyzer analyzes number frequency in historical draws
type FrequencyAnalyzer struct {
	name     string
	weight   float64
	minDraws int
	mu       sync.RWMutex
}

// NewFrequencyAnalyzer creates a new frequency analyzer
func NewFrequencyAnalyzer(weight float64) *FrequencyAnalyzer {
	return &FrequencyAnalyzer{
		name:     "frequency_analysis",
		weight:   weight,
		minDraws: 8, // Minimum 8 draws needed for frequency analysis
	}
}

// Name returns the algorithm name
func (fa *FrequencyAnalyzer) Name() string {
	return fa.name
}

// GetWeight returns the algorithm's weight
func (fa *FrequencyAnalyzer) GetWeight() float64 {
	fa.mu.RLock()
	defer fa.mu.RUnlock()
	return fa.weight
}

// SetWeight sets the algorithm's weight
func (fa *FrequencyAnalyzer) SetWeight(weight float64) error {
	if weight < 0 {
		return fmt.Errorf("weight cannot be negative, got %f", weight)
	}
	fa.mu.Lock()
	defer fa.mu.Unlock()
	fa.weight = weight
	return nil
}

// Validate checks if there's enough data for prediction
func (fa *FrequencyAnalyzer) Validate(historicalData []*entity.Draw) error {
	if len(historicalData) < fa.minDraws {
		return fmt.Errorf("need at least %d draws for frequency analysis, got %d",
			fa.minDraws, len(historicalData))
	}
	return nil
}

// Train updates algorithm parameters (frequency analyzer doesn't need training)
func (fa *FrequencyAnalyzer) Train(ctx context.Context, historicalData []*entity.Draw) error {
	// Frequency analyzer doesn't require training
	return nil
}

// Predict generates predictions based on number frequency
func (fa *FrequencyAnalyzer) Predict(
	ctx context.Context,
	gameType valueobject.GameType,
	historicalData []*entity.Draw,
) (*entity.Prediction, error) {
	// Validate input
	if err := fa.Validate(historicalData); err != nil {
		return nil, err
	}

	// Check for cancellation
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// Get number range for game type
	minRange, maxRange := gameType.NumberRange()

	// Count frequency of each number
	frequency := make(map[int]int)
	totalNumbers := 0

	for _, draw := range historicalData {
		for _, num := range draw.Numbers {
			frequency[num]++
			totalNumbers++
		}
	}

	// Calculate expected frequency and variance
	expectedFreq := float64(totalNumbers) / float64((maxRange-minRange+1)*len(historicalData))

	// Create number-frequency pairs
	type numFreq struct {
		num   int
		count int
		score float64
	}

	pairs := make([]numFreq, 0)
	for i := minRange; i <= maxRange; i++ {
		count := frequency[i]
		// Score is based on how much the frequency exceeds expected
		score := float64(count) / expectedFreq
		pairs = append(pairs, numFreq{
			num:   i,
			count: count,
			score: score,
		})
	}

	// Sort by frequency score (descending), then by frequency
	sort.Slice(pairs, func(i, j int) bool {
		if pairs[i].score != pairs[j].score {
			return pairs[i].score > pairs[j].score
		}
		return pairs[i].count > pairs[j].count
	})

	// Take top 6 most frequent numbers
	predictedNums := make([]int, 6)
	for i := 0; i < 6; i++ {
		predictedNums[i] = pairs[i].num
	}

	// Create numbers value object
	numbers, err := valueobject.NewNumbers(predictedNums)
	if err != nil {
		return nil, fmt.Errorf("failed to create numbers: %w", err)
	}

	// Calculate confidence based on frequency consistency
	confidence := fa.calculateConfidence(frequency, numbers, expectedFreq)

	// Create prediction
	prediction := &entity.Prediction{
		ID:            "", // Will be set by repository
		GameType:      gameType,
		AlgorithmName: fa.name,
		Numbers:       numbers,
		Confidence:    confidence,
		GeneratedAt:   time.Now(),
		ForDate:       time.Now().Add(24 * time.Hour), // Predict for tomorrow
		Metadata: map[string]string{
			"min_draws_required": fmt.Sprintf("%d", fa.minDraws),
			"total_draws_used":   fmt.Sprintf("%d", len(historicalData)),
			"expected_freq":      fmt.Sprintf("%.4f", expectedFreq),
		},
	}

	return prediction, nil
}

// calculateConfidence calculates prediction confidence
func (fa *FrequencyAnalyzer) calculateConfidence(
	frequency map[int]int,
	numbers valueobject.Numbers,
	expectedFreq float64,
) float64 {
	// Calculate average relative frequency of selected numbers
	totalScore := 0.0
	for _, num := range numbers {
		score := float64(frequency[num]) / expectedFreq
		totalScore += score
	}
	avgScore := totalScore / 6.0

	// Normalize to 0-1 range (assuming max reasonable score is 2.0)
	confidence := avgScore / 2.0
	if confidence > 1.0 {
		confidence = 1.0
	}
	if confidence < 0.1 {
		confidence = 0.1
	}

	return confidence
}

// SetMinDraws sets the minimum number of draws required for prediction
func (fa *FrequencyAnalyzer) SetMinDraws(minDraws int) error {
	if minDraws < 10 {
		return fmt.Errorf("minimum draws must be at least 10, got %d", minDraws)
	}
	fa.mu.Lock()
	defer fa.mu.Unlock()
	fa.minDraws = minDraws
	return nil
}

// GetMinDraws returns the minimum number of draws required
func (fa *FrequencyAnalyzer) GetMinDraws() int {
	fa.mu.RLock()
	defer fa.mu.RUnlock()
	return fa.minDraws
}
