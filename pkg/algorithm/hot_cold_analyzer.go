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

// HotColdAnalyzer identifies hot (recently drawn) and cold (overdue) numbers
type HotColdAnalyzer struct {
	name          string
	weight        float64
	minDraws      int
	hotThreshold  int // Number of recent draws to consider for "hot" numbers
	coldThreshold int // Number of draws since last appearance for "cold" numbers
	mu            sync.RWMutex
}

// NewHotColdAnalyzer creates a new hot/cold analyzer
func NewHotColdAnalyzer(weight float64) *HotColdAnalyzer {
	return &HotColdAnalyzer{
		name:          "hot_cold_analysis",
		weight:        weight,
		minDraws:      50,
		hotThreshold:  20,
		coldThreshold: 15,
	}
}

// Name returns the algorithm name
func (hca *HotColdAnalyzer) Name() string {
	return hca.name
}

// GetWeight returns the algorithm's weight
func (hca *HotColdAnalyzer) GetWeight() float64 {
	hca.mu.RLock()
	defer hca.mu.RUnlock()
	return hca.weight
}

// SetWeight sets the algorithm's weight
func (hca *HotColdAnalyzer) SetWeight(weight float64) error {
	if weight < 0 {
		return fmt.Errorf("weight cannot be negative, got %f", weight)
	}
	hca.mu.Lock()
	defer hca.mu.Unlock()
	hca.weight = weight
	return nil
}

// Validate checks if there's enough data for prediction
func (hca *HotColdAnalyzer) Validate(historicalData []*entity.Draw) error {
	if len(historicalData) < hca.minDraws {
		return fmt.Errorf("need at least %d draws for hot/cold analysis, got %d",
			hca.minDraws, len(historicalData))
	}
	return nil
}

// Train updates algorithm parameters (hot/cold doesn't need training)
func (hca *HotColdAnalyzer) Train(ctx context.Context, historicalData []*entity.Draw) error {
	return nil
}

// Predict generates predictions based on hot and cold number analysis
func (hca *HotColdAnalyzer) Predict(
	ctx context.Context,
	gameType valueobject.GameType,
	historicalData []*entity.Draw,
) (*entity.Prediction, error) {
	if err := hca.Validate(historicalData); err != nil {
		return nil, err
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	hca.mu.RLock()
	hotThreshold := hca.hotThreshold
	coldThreshold := hca.coldThreshold
	hca.mu.RUnlock()

	// Reverse to get most recent first
	recentDraws := reverseDraws(historicalData)

	// Find hot numbers (frequently drawn in recent draws)
	hotNumbers := hca.findHotNumbers(recentDraws, hotThreshold, gameType)

	// Find cold numbers (haven't been drawn recently)
	coldNumbers := hca.findColdNumbers(recentDraws, coldThreshold, gameType)

	// Combine: 3 hot + 3 cold numbers
	predictedNums := append(hotNumbers[:3], coldNumbers[:3]...)
	sort.Ints(predictedNums)

	// Validate and create numbers
	numbers, err := valueobject.NewNumbers(predictedNums)
	if err != nil {
		return nil, fmt.Errorf("failed to create numbers: %w", err)
	}

	// Calculate confidence
	confidence := hca.calculateConfidence(hotNumbers, coldNumbers)

	prediction := &entity.Prediction{
		ID:            "",
		GameType:      gameType,
		AlgorithmName: hca.name,
		Numbers:       numbers,
		Confidence:    confidence,
		GeneratedAt:   time.Now(),
		ForDate:       time.Now().Add(24 * time.Hour),
		Metadata: map[string]string{
			"hot_threshold":  fmt.Sprintf("%d", hotThreshold),
			"cold_threshold": fmt.Sprintf("%d", coldThreshold),
			"hot_numbers":    fmt.Sprintf("%v", hotNumbers),
			"cold_numbers":   fmt.Sprintf("%v", coldNumbers),
		},
	}

	return prediction, nil
}

// findHotNumbers identifies numbers that have appeared frequently in recent draws
func (hca *HotColdAnalyzer) findHotNumbers(
	draws []*entity.Draw,
	limit int,
	gameType valueobject.GameType,
) []int {
	minRange, maxRange := gameType.NumberRange()

	// Count frequency in recent draws
	frequency := make(map[int]int)
	drawsToCheck := limit
	if drawsToCheck > len(draws) {
		drawsToCheck = len(draws)
	}

	for i := 0; i < drawsToCheck; i++ {
		for _, num := range draws[i].Numbers {
			frequency[num]++
		}
	}

	// Sort by frequency
	type numFreq struct {
		num   int
		count int
	}

	sorted := make([]numFreq, 0)
	for num := minRange; num <= maxRange; num++ {
		sorted = append(sorted, numFreq{
			num:   num,
			count: frequency[num],
		})
	}

	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].count > sorted[j].count
	})

	// Return top hot numbers (more than we need)
	result := make([]int, 0, 5)
	for i := 0; i < len(sorted) && i < 5; i++ {
		result = append(result, sorted[i].num)
	}

	return result
}

// findColdNumbers identifies numbers that haven't appeared recently
func (hca *HotColdAnalyzer) findColdNumbers(
	draws []*entity.Draw,
	threshold int,
	gameType valueobject.GameType,
) []int {
	minRange, maxRange := gameType.NumberRange()

	// Track last appearance of each number
	lastSeen := make(map[int]int) // number -> index (0 = most recent)

	for i, draw := range draws {
		for _, num := range draw.Numbers {
			if _, exists := lastSeen[num]; !exists {
				lastSeen[num] = i
			}
		}
	}

	// Find cold numbers (not seen in threshold draws)
	coldNumbers := make([]int, 0)
	for num := minRange; num <= maxRange; num++ {
		if lastSeen, exists := lastSeen[num]; !exists || lastSeen >= threshold {
			coldNumbers = append(coldNumbers, num)
		}
	}

	// If we don't have enough cold numbers, lower threshold
	if len(coldNumbers) < 3 {
		coldNumbers = make([]int, 0)
		newThreshold := threshold / 2
		for num := minRange; num <= maxRange; num++ {
			if lastSeen, exists := lastSeen[num]; !exists || lastSeen >= newThreshold {
				coldNumbers = append(coldNumbers, num)
			}
		}
	}

	// Sort by how long they've been cold (ascending index = colder)
	sort.Slice(coldNumbers, func(i, j int) bool {
		lastI, existsI := lastSeen[coldNumbers[i]]
		lastJ, existsJ := lastSeen[coldNumbers[j]]
		if !existsI && existsJ {
			return true
		}
		if existsI && existsJ {
			return lastI > lastJ
		}
		return false
	})

	// Return top 5 cold numbers
	if len(coldNumbers) > 5 {
		coldNumbers = coldNumbers[:5]
	}

	return coldNumbers
}

// calculateConfidence calculates prediction confidence
func (hca *HotColdAnalyzer) calculateConfidence(
	hotNumbers []int,
	coldNumbers []int,
) float64 {
	// Confidence based on balance of hot and cold
	// Ideally, we want a good mix (not too many extremes)

	hotCount := len(hotNumbers)
	coldCount := len(coldNumbers)

	// Higher confidence when we have both hot and cold numbers
	if hotCount > 0 && coldCount > 0 {
		confidence := 0.7
		if hotCount >= 3 && coldCount >= 3 {
			confidence = 0.85
		}
		return confidence
	}

	// Lower confidence if we only have one type
	return 0.5
}

// reverseDraws reverses the order of draws (most recent first)
func reverseDraws(draws []*entity.Draw) []*entity.Draw {
	reversed := make([]*entity.Draw, len(draws))
	for i, draw := range draws {
		reversed[len(draws)-1-i] = draw
	}
	return reversed
}

// SetHotThreshold sets the threshold for hot number detection
func (hca *HotColdAnalyzer) SetHotThreshold(threshold int) error {
	if threshold < 5 {
		return fmt.Errorf("hot threshold must be at least 5, got %d", threshold)
	}
	hca.mu.Lock()
	defer hca.mu.Unlock()
	hca.hotThreshold = threshold
	return nil
}

// SetColdThreshold sets the threshold for cold number detection
func (hca *HotColdAnalyzer) SetColdThreshold(threshold int) error {
	if threshold < 5 {
		return fmt.Errorf("cold threshold must be at least 5, got %d", threshold)
	}
	hca.mu.Lock()
	defer hca.mu.Unlock()
	hca.coldThreshold = threshold
	return nil
}

// GetHotThreshold returns the hot threshold
func (hca *HotColdAnalyzer) GetHotThreshold() int {
	hca.mu.RLock()
	defer hca.mu.RUnlock()
	return hca.hotThreshold
}

// GetColdThreshold returns the cold threshold
func (hca *HotColdAnalyzer) GetColdThreshold() int {
	hca.mu.RLock()
	defer hca.mu.RUnlock()
	return hca.coldThreshold
}
