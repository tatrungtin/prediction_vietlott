package algorithm

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/tool_predict/internal/domain/entity"
	"github.com/tool_predict/internal/domain/valueobject"
)

// PatternAnalyzer analyzes various patterns in lottery numbers
type PatternAnalyzer struct {
	name     string
	weight   float64
	minDraws int
	mu       sync.RWMutex
}

// NewPatternAnalyzer creates a new pattern analyzer
func NewPatternAnalyzer(weight float64) *PatternAnalyzer {
	return &PatternAnalyzer{
		name:     "pattern_analysis",
		weight:   weight,
		minDraws: 100,
	}
}

// Name returns the algorithm name
func (pa *PatternAnalyzer) Name() string {
	return pa.name
}

// GetWeight returns the algorithm's weight
func (pa *PatternAnalyzer) GetWeight() float64 {
	pa.mu.RLock()
	defer pa.mu.RUnlock()
	return pa.weight
}

// SetWeight sets the algorithm's weight
func (pa *PatternAnalyzer) SetWeight(weight float64) error {
	if weight < 0 {
		return fmt.Errorf("weight cannot be negative, got %f", weight)
	}
	pa.mu.Lock()
	defer pa.mu.Unlock()
	pa.weight = weight
	return nil
}

// Validate checks if there's enough data for prediction
func (pa *PatternAnalyzer) Validate(historicalData []*entity.Draw) error {
	if len(historicalData) < pa.minDraws {
		return fmt.Errorf("need at least %d draws for pattern analysis, got %d",
			pa.minDraws, len(historicalData))
	}
	return nil
}

// Train updates algorithm parameters (pattern analyzer doesn't need training)
func (pa *PatternAnalyzer) Train(ctx context.Context, historicalData []*entity.Draw) error {
	return nil
}

// Predict generates predictions based on pattern analysis
func (pa *PatternAnalyzer) Predict(
	ctx context.Context,
	gameType valueobject.GameType,
	historicalData []*entity.Draw,
) (*entity.Prediction, error) {
	if err := pa.Validate(historicalData); err != nil {
		return nil, err
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// Analyze multiple patterns
	consecutivePattern := pa.analyzeConsecutiveNumbers(historicalData)
	oddEvenPattern := pa.analyzeOddEvenRatio(historicalData)
	sumPattern := pa.analyzeSumRanges(historicalData, gameType)
	lowHighPattern := pa.analyzeLowHighRatio(historicalData, gameType)

	// Combine patterns to generate prediction
	predictedNums := pa.combinePatterns(
		consecutivePattern,
		oddEvenPattern,
		sumPattern,
		lowHighPattern,
		gameType,
	)

	sort.Ints(predictedNums)

	numbers, err := valueobject.NewNumbers(predictedNums)
	if err != nil {
		return nil, fmt.Errorf("failed to create numbers: %w", err)
	}

	confidence := pa.calculateConfidence(historicalData)

	prediction := &entity.Prediction{
		ID:            "",
		GameType:      gameType,
		AlgorithmName: pa.name,
		Numbers:       numbers,
		Confidence:    confidence,
		GeneratedAt:   time.Now(),
		ForDate:       time.Now().Add(24 * time.Hour),
		Metadata: map[string]string{
			"consecutive_pairs": strings.Trim(strings.Join(fmtIntSlice(consecutivePattern), ","), "[]"),
			"target_odd_count":  fmt.Sprintf("%d", oddEvenPattern.targetOddCount),
			"sum_range":         fmt.Sprintf("%d-%d", sumPattern.minSum, sumPattern.maxSum),
			"low_high_ratio":    fmt.Sprintf("%.2f", lowHighPattern.ratio),
		},
	}

	return prediction, nil
}

// analyzeConsecutiveNumbers finds pairs that frequently appear together
func (pa *PatternAnalyzer) analyzeConsecutiveNumbers(draws []*entity.Draw) []int {
	pairCount := make(map[string]int)

	for _, draw := range draws {
		nums := draw.Numbers
		for i := 0; i < len(nums)-1; i++ {
			if nums[i+1]-nums[i] == 1 {
				pair := fmt.Sprintf("%d-%d", nums[i], nums[i+1])
				pairCount[pair]++
			}
		}
	}

	// Find frequently consecutive pairs
	consecutiveNumbers := make(map[int]bool)
	threshold := len(draws) / 20 // At least 5% of draws
	if threshold < 2 {
		threshold = 2
	}

	for pair, count := range pairCount {
		if count >= threshold {
			parts := strings.Split(pair, "-")
			n1, n2 := parseInt(parts[0]), parseInt(parts[1])
			consecutiveNumbers[n1] = true
			consecutiveNumbers[n2] = true
		}
	}

	result := make([]int, 0, len(consecutiveNumbers))
	for num := range consecutiveNumbers {
		result = append(result, num)
	}

	return result
}

// oddEvenPattern represents the ideal odd/even distribution
type oddEvenPattern struct {
	targetOddCount  int
	targetEvenCount int
}

// analyzeOddEvenRatio analyzes the odd/even number distribution
func (pa *PatternAnalyzer) analyzeOddEvenRatio(draws []*entity.Draw) oddEvenPattern {
	totalOdd := 0

	for _, draw := range draws {
		for _, num := range draw.Numbers {
			if num%2 == 1 {
				totalOdd++
			}
		}
	}

	avgOdd := float64(totalOdd) / float64(len(draws)*6)

	// Convert to actual counts (must sum to 6)
	targetOdd := int(math.Round(avgOdd))
	if targetOdd < 2 {
		targetOdd = 2
	}
	if targetOdd > 4 {
		targetOdd = 4
	}

	targetEven := 6 - targetOdd

	return oddEvenPattern{
		targetOddCount:  targetOdd,
		targetEvenCount: targetEven,
	}
}

// sumPattern represents the typical sum range
type sumPattern struct {
	minSum int
	maxSum int
}

// analyzeSumRanges analyzes the sum of numbers in each draw
func (pa *PatternAnalyzer) analyzeSumRanges(draws []*entity.Draw, gameType valueobject.GameType) sumPattern {
	sums := make([]int, len(draws))

	for i, draw := range draws {
		sums[i] = draw.Numbers.Sum()
	}

	// Calculate statistics
	mean := calculateMean(sums)
	stdDev := calculateStdDev(sums, mean)

	// Typical range is mean ± 1 stdDev
	minSum := int(mean - stdDev)
	maxSum := int(mean + stdDev)

	// Ensure reasonable bounds
	minRange, maxRange := gameType.NumberRange()
	minPossibleSum := minRange + (minRange + 1) + (minRange + 2) + (minRange + 3) + (minRange + 4) + (minRange + 5)
	maxPossibleSum := maxRange + (maxRange - 1) + (maxRange - 2) + (maxRange - 3) + (maxRange - 4) + (maxRange - 5)

	if minSum < minPossibleSum {
		minSum = minPossibleSum
	}
	if maxSum > maxPossibleSum {
		maxSum = maxPossibleSum
	}

	return sumPattern{
		minSum: minSum,
		maxSum: maxSum,
	}
}

// lowHighPattern represents the low/high number ratio
type lowHighPattern struct {
	ratio float64
}

// analyzeLowHighRatio analyzes the ratio of low to high numbers
func (pa *PatternAnalyzer) analyzeLowHighRatio(draws []*entity.Draw, gameType valueobject.GameType) lowHighPattern {
	minRange, maxRange := gameType.NumberRange()
	midpoint := (minRange + maxRange) / 2

	totalLow := 0
	totalNumbers := 0

	for _, draw := range draws {
		for _, num := range draw.Numbers {
			if num <= midpoint {
				totalLow++
			}
			totalNumbers++
		}
	}

	ratio := float64(totalLow) / float64(totalNumbers)

	return lowHighPattern{
		ratio: ratio,
	}
}

// combinePatterns combines all pattern analyses to generate a prediction
func (pa *PatternAnalyzer) combinePatterns(
	consecutivePattern []int,
	oddEvenPattern oddEvenPattern,
	sumPattern sumPattern,
	lowHighPattern lowHighPattern,
	gameType valueobject.GameType,
) []int {
	minRange, maxRange := gameType.NumberRange()

	// Start with consecutive pattern numbers
	selectedNumbers := make(map[int]bool)
	for _, num := range consecutivePattern {
		if num >= minRange && num <= maxRange {
			selectedNumbers[num] = true
		}
	}

	// Fill remaining to match odd/even pattern
	oddCount := 0
	evenCount := 0
	for num := range selectedNumbers {
		if num%2 == 0 {
			evenCount++
		} else {
			oddCount++
		}
	}

	// Add numbers to achieve target odd/even distribution
	neededOdd := oddEvenPattern.targetOddCount - oddCount
	neededEven := oddEvenPattern.targetEvenCount - evenCount

	// Priority: use consecutive numbers first, then fill to match patterns
	additionalNumbers := make([]int, 0)

	// Add odd numbers if needed
	if neededOdd > 0 {
		for i := minRange; i <= maxRange && neededOdd > 0; i++ {
			if !selectedNumbers[i] && i%2 == 1 {
				// Check if this fits the sum pattern later
				additionalNumbers = append(additionalNumbers, i)
				selectedNumbers[i] = true
				neededOdd--
			}
		}
	}

	// Add even numbers if needed
	if neededEven > 0 {
		for i := minRange; i <= maxRange && neededEven > 0; i++ {
			if !selectedNumbers[i] && i%2 == 0 {
				additionalNumbers = append(additionalNumbers, i)
				selectedNumbers[i] = true
				neededEven--
			}
		}
	}

	// Convert to slice and sort
	result := make([]int, 0, len(selectedNumbers))
	for num := range selectedNumbers {
		result = append(result, num)
	}

	// Ensure we have exactly 6 numbers first
	if len(result) > 6 {
		// Keep the first 6 after sorting
		sort.Ints(result)
		result = result[:6]
	} else if len(result) < 6 {
		// Add more numbers to reach 6
		result = pa.fillToSix(result, gameType)
	}

	// Adjust to fit sum range if needed (only after we have exactly 6 numbers)
	currentSum := sumIntSlice(result)
	if currentSum < sumPattern.minSum || currentSum > sumPattern.maxSum {
		result = pa.adjustForSumRange(result, sumPattern, gameType)
	}

	sort.Ints(result)
	return result
}

// adjustForSumRange adjusts numbers to fit the target sum range
func (pa *PatternAnalyzer) adjustForSumRange(numbers []int, sumPattern sumPattern, gameType valueobject.GameType) []int {
	minRange, maxRange := gameType.NumberRange()
	currentSum := sumIntSlice(numbers)
	targetSum := (sumPattern.minSum + sumPattern.maxSum) / 2

	adjustment := targetSum - currentSum

	// Try to adjust by swapping numbers
	result := make([]int, len(numbers))
	copy(result, numbers)

	for i := 0; i < len(result) && adjustment != 0; i++ {
		if adjustment > 0 {
			// Need to increase sum
			if result[i] < maxRange {
				result[i]++
				adjustment--
			}
		} else {
			// Need to decrease sum
			if result[i] > minRange {
				result[i]--
				adjustment++
			}
		}
	}

	return result
}

// fillToSix fills the array to have exactly 6 numbers
func (pa *PatternAnalyzer) fillToSix(numbers []int, gameType valueobject.GameType) []int {
	minRange, maxRange := gameType.NumberRange()
	selected := make(map[int]bool)
	for _, num := range numbers {
		selected[num] = true
	}

	// Add missing numbers
	for i := minRange; i <= maxRange && len(numbers) < 6; i++ {
		if !selected[i] {
			numbers = append(numbers, i)
			selected[i] = true
		}
	}

	return numbers
}

// calculateConfidence calculates prediction confidence
func (pa *PatternAnalyzer) calculateConfidence(historicalData []*entity.Draw) float64 {
	// Confidence based on pattern consistency
	// Start with base confidence
	confidence := 0.65

	// Increase confidence with more data
	if len(historicalData) >= 200 {
		confidence = 0.75
	}

	return confidence
}

// Helper functions

func calculateMean(values []int) float64 {
	if len(values) == 0 {
		return 0
	}
	sum := 0
	for _, v := range values {
		sum += v
	}
	return float64(sum) / float64(len(values))
}

func calculateStdDev(values []int, mean float64) float64 {
	if len(values) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range values {
		diff := float64(v) - mean
		sum += diff * diff
	}
	variance := sum / float64(len(values))
	return math.Sqrt(variance)
}

func sumIntSlice(values []int) int {
	sum := 0
	for _, v := range values {
		sum += v
	}
	return sum
}

func fmtIntSlice(values []int) []string {
	result := make([]string, len(values))
	for i, v := range values {
		result[i] = fmt.Sprintf("%d", v)
	}
	return result
}

func parseInt(s string) int {
	result := 0
	for _, ch := range s {
		result = result*10 + int(ch-'0')
	}
	return result
}
