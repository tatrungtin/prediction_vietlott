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

// VotingStrategy defines the type of voting to use
type VotingStrategy string

const (
	WeightedVoting      VotingStrategy = "weighted"
	MajorityVoting      VotingStrategy = "majority"
	ConfidenceWeighted  VotingStrategy = "confidence_weighted"
)

// Ensemble combines multiple algorithms using voting strategies
type Ensemble struct {
	registry       *Registry
	votingStrategy VotingStrategy
	mu             sync.RWMutex
}

// NewEnsemble creates a new ensemble with the given registry and voting strategy
func NewEnsemble(registry *Registry, votingStrategy VotingStrategy) *Ensemble {
	return &Ensemble{
		registry:       registry,
		votingStrategy: votingStrategy,
	}
}

// SetVotingStrategy changes the voting strategy
func (e *Ensemble) SetVotingStrategy(strategy VotingStrategy) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.votingStrategy = strategy
}

// GetVotingStrategy returns the current voting strategy
func (e *Ensemble) GetVotingStrategy() VotingStrategy {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.votingStrategy
}

// GeneratePredictions generates predictions from all algorithms and combines them
func (e *Ensemble) GeneratePredictions(
	ctx context.Context,
	gameType valueobject.GameType,
	historicalData []*entity.Draw,
) (*entity.EnsemblePrediction, error) {
	algorithms := e.registry.GetAll()

	if len(algorithms) == 0 {
		return nil, fmt.Errorf("no algorithms registered in the ensemble")
	}

	// Generate predictions from all algorithms
	predictions := make([]*entity.Prediction, 0, len(algorithms))
	for _, algo := range algorithms {
		if err := algo.Validate(historicalData); err != nil {
			// Skip algorithms that can't predict
			continue
		}

		pred, err := algo.Predict(ctx, gameType, historicalData)
		if err != nil {
			// Log error but continue with other algorithms
			continue
		}

		predictions = append(predictions, pred)
	}

	if len(predictions) == 0 {
		return nil, fmt.Errorf("no valid predictions generated from any algorithm")
	}

	// Apply voting strategy
	e.mu.RLock()
	strategy := e.votingStrategy
	e.mu.RUnlock()

	finalNumbers, err := e.applyVotingStrategy(predictions, strategy)
	if err != nil {
		return nil, fmt.Errorf("failed to apply voting strategy: %w", err)
	}

	// Calculate algorithm contributions
	contributions := e.calculateContributions(predictions, finalNumbers)

	// Create ensemble prediction
	ensemblePred := &entity.EnsemblePrediction{
		ID:             "", // Will be set by repository
		GameType:       gameType,
		Predictions:    predictions,
		FinalNumbers:   finalNumbers,
		VotingStrategy: string(strategy),
		GeneratedAt:    time.Now(),
		AlgorithmStats: contributions,
	}

	return ensemblePred, nil
}

// applyVotingStrategy applies the specified voting strategy
func (e *Ensemble) applyVotingStrategy(
	predictions []*entity.Prediction,
	strategy VotingStrategy,
) (valueobject.Numbers, error) {
	switch strategy {
	case WeightedVoting:
		return e.weightedVoting(predictions)
	case MajorityVoting:
		return e.majorityVoting(predictions)
	case ConfidenceWeighted:
		return e.confidenceWeightedVoting(predictions)
	default:
		return e.weightedVoting(predictions)
	}
}

// weightedVoting uses algorithm weights from the registry for voting
func (e *Ensemble) weightedVoting(predictions []*entity.Prediction) (valueobject.Numbers, error) {
	voteCount := make(map[int]float64)

	for _, pred := range predictions {
		weight := e.registry.GetWeight(pred.AlgorithmName)
		for _, num := range pred.Numbers {
			voteCount[num] += weight
		}
	}

	// Sort by vote count
	type numVote struct {
		num   int
		votes float64
	}

	sorted := make([]numVote, 0, len(voteCount))
	for num, votes := range voteCount {
		sorted = append(sorted, numVote{num: num, votes: votes})
	}

	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].votes > sorted[j].votes
	})

	// Take top 6
	result := make([]int, 6)
	for i := 0; i < 6 && i < len(sorted); i++ {
		result[i] = sorted[i].num
	}

	// Handle ties - if we have less than 6, add more
	if len(sorted) < 6 {
		// This is rare, but handle it by adding from predictions
		result = e.fillRemainingFromPredictions(result, predictions)
	}

	sort.Ints(result)
	return valueobject.NewNumbers(result)
}

// majorityVoting uses simple majority voting
func (e *Ensemble) majorityVoting(predictions []*entity.Prediction) (valueobject.Numbers, error) {
	voteCount := make(map[int]int)

	for _, pred := range predictions {
		for _, num := range pred.Numbers {
			voteCount[num]++
		}
	}

	// Sort by vote count
	type numVote struct {
		num   int
		votes int
	}

	sorted := make([]numVote, 0, len(voteCount))
	for num, votes := range voteCount {
		sorted = append(sorted, numVote{num: num, votes: votes})
	}

	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].votes > sorted[j].votes
	})

	// Take top 6
	result := make([]int, 6)
	for i := 0; i < 6 && i < len(sorted); i++ {
		result[i] = sorted[i].num
	}

	sort.Ints(result)
	return valueobject.NewNumbers(result)
}

// confidenceWeightedVoting uses confidence scores as weights
func (e *Ensemble) confidenceWeightedVoting(predictions []*entity.Prediction) (valueobject.Numbers, error) {
	voteCount := make(map[int]float64)

	for _, pred := range predictions {
		for _, num := range pred.Numbers {
			voteCount[num] += pred.Confidence
		}
	}

	// Sort by vote count
	type numVote struct {
		num   int
		votes float64
	}

	sorted := make([]numVote, 0, len(voteCount))
	for num, votes := range voteCount {
		sorted = append(sorted, numVote{num: num, votes: votes})
	}

	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].votes > sorted[j].votes
	})

	// Take top 6
	result := make([]int, 6)
	for i := 0; i < 6 && i < len(sorted); i++ {
		result[i] = sorted[i].num
	}

	sort.Ints(result)
	return valueobject.NewNumbers(result)
}

// fillRemainingFromPredictions fills remaining slots from predictions
func (e *Ensemble) fillRemainingFromPredictions(
	current []int,
	predictions []*entity.Prediction,
) []int {
	used := make(map[int]bool)
	for _, num := range current {
		if num != 0 {
			used[num] = true
		}
	}

	result := make([]int, 0, 6)
	for _, num := range current {
		if num != 0 {
			result = append(result, num)
		}
	}

	// Fill remaining slots from predictions
	for _, pred := range predictions {
		for _, num := range pred.Numbers {
			if !used[num] && len(result) < 6 {
				result = append(result, num)
				used[num] = true
			}
		}
		if len(result) >= 6 {
			break
		}
	}

	return result
}

// calculateContributions calculates each algorithm's contribution to the final result
func (e *Ensemble) calculateContributions(
	predictions []*entity.Prediction,
	finalNumbers valueobject.Numbers,
) []entity.AlgorithmContribution {
	contributions := make([]entity.AlgorithmContribution, 0, len(predictions))

	for _, pred := range predictions {
		matchCount := finalNumbers.MatchCount(pred.Numbers)
		contribution := entity.AlgorithmContribution{
			AlgorithmName: pred.AlgorithmName,
			Weight:        e.registry.GetWeight(pred.AlgorithmName),
			MatchCount:    matchCount,
			Confidence:    pred.Confidence,
		}
		contributions = append(contributions, contribution)
	}

	return contributions
}

// GetBestAlgorithm returns the algorithm that contributed most to the ensemble
func (e *Ensemble) GetBestAlgorithm(contributions []entity.AlgorithmContribution) *entity.AlgorithmContribution {
	if len(contributions) == 0 {
		return nil
	}

	best := &contributions[0]
	for i := 1; i < len(contributions); i++ {
		// Compare by match count first, then by confidence
		if contributions[i].MatchCount > best.MatchCount ||
			(contributions[i].MatchCount == best.MatchCount && contributions[i].Confidence > best.Confidence) {
			best = &contributions[i]
		}
	}

	return best
}

// UpdateWeights updates algorithm weights based on their performance
func (e *Ensemble) UpdateWeights(performanceStats map[string]float64) error {
	for algoName, performance := range performanceStats {
		// Update weight based on performance (0-1 range)
		newWeight := performance
		if newWeight < 0.1 {
			newWeight = 0.1 // Minimum weight
		}
		if newWeight > 2.0 {
			newWeight = 2.0 // Maximum weight
		}

		if err := e.registry.UpdateWeight(algoName, newWeight); err != nil {
			return fmt.Errorf("failed to update weight for %s: %w", algoName, err)
		}
	}
	return nil
}

// GetConsensusScore returns a score indicating how much the algorithms agree
// 1.0 = perfect agreement, 0.0 = no agreement
func (e *Ensemble) GetConsensusScore(predictions []*entity.Prediction) float64 {
	if len(predictions) < 2 {
		return 1.0
	}

	// Calculate pairwise similarity
	totalSimilarity := 0.0
	comparisons := 0

	for i := 0; i < len(predictions); i++ {
		for j := i + 1; j < len(predictions); j++ {
			matchCount := predictions[i].Numbers.MatchCount(predictions[j].Numbers)
			similarity := float64(matchCount) / 6.0
			totalSimilarity += similarity
			comparisons++
		}
	}

	if comparisons == 0 {
		return 1.0
	}

	return totalSimilarity / float64(comparisons)
}
