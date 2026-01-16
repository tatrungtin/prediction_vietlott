package algorithm

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tool_predict/internal/domain/valueobject"
)

func TestRegistry_Register(t *testing.T) {
	registry := NewRegistry()
	analyzer := NewFrequencyAnalyzer(1.0)

	err := registry.Register(analyzer, 1.0)
	require.NoError(t, err)
	assert.Equal(t, 1, registry.Count())

	// Test duplicate registration
	err = registry.Register(analyzer, 1.5)
	assert.Error(t, err)
	assert.Equal(t, 1, registry.Count())
}

func TestRegistry_RegisterOrUpdate(t *testing.T) {
	registry := NewRegistry()
	analyzer := NewFrequencyAnalyzer(1.0)

	// First registration
	err := registry.RegisterOrUpdate(analyzer, 1.0)
	require.NoError(t, err)
	assert.Equal(t, 1, registry.Count())

	// Update existing
	err = registry.RegisterOrUpdate(analyzer, 1.5)
	require.NoError(t, err)
	assert.Equal(t, 1, registry.Count())

	// Verify weight was updated
	weight := registry.GetWeight("frequency_analysis")
	assert.Equal(t, 1.5, weight)
}

func TestRegistry_Get(t *testing.T) {
	registry := NewRegistry()
	analyzer := NewFrequencyAnalyzer(1.0)

	// Test getting non-existent algorithm
	_, err := registry.Get("nonexistent")
	assert.Error(t, err)

	// Register and get
	err = registry.Register(analyzer, 1.0)
	require.NoError(t, err)

	retrieved, err := registry.Get("frequency_analysis")
	require.NoError(t, err)
	assert.Equal(t, analyzer, retrieved)
}

func TestRegistry_GetAll(t *testing.T) {
	registry := NewRegistry()

	// Empty registry
	algorithms := registry.GetAll()
	assert.Equal(t, 0, len(algorithms))

	// Add algorithms
	analyzer1 := NewFrequencyAnalyzer(1.0)
	analyzer2 := NewHotColdAnalyzer(1.2)
	analyzer3 := NewPatternAnalyzer(0.8)

	require.NoError(t, registry.Register(analyzer1, 1.0))
	require.NoError(t, registry.Register(analyzer2, 1.2))
	require.NoError(t, registry.Register(analyzer3, 0.8))

	algorithms = registry.GetAll()
	assert.Equal(t, 3, len(algorithms))
}

func TestRegistry_GetNames(t *testing.T) {
	registry := NewRegistry()

	analyzer1 := NewFrequencyAnalyzer(1.0)
	analyzer2 := NewHotColdAnalyzer(1.2)

	require.NoError(t, registry.Register(analyzer1, 1.0))
	require.NoError(t, registry.Register(analyzer2, 1.2))

	names := registry.GetNames()
	assert.Equal(t, 2, len(names))
	assert.Contains(t, names, "frequency_analysis")
	assert.Contains(t, names, "hot_cold_analysis")
}

func TestRegistry_UpdateWeight(t *testing.T) {
	registry := NewRegistry()
	analyzer := NewFrequencyAnalyzer(1.0)

	require.NoError(t, registry.Register(analyzer, 1.0))

	err := registry.UpdateWeight("frequency_analysis", 2.0)
	require.NoError(t, err)

	weight := registry.GetWeight("frequency_analysis")
	assert.Equal(t, 2.0, weight)

	// Test updating non-existent algorithm
	err = registry.UpdateWeight("nonexistent", 1.0)
	assert.Error(t, err)

	// Test negative weight
	err = registry.UpdateWeight("frequency_analysis", -1.0)
	assert.Error(t, err)
}

func TestRegistry_Unregister(t *testing.T) {
	registry := NewRegistry()
	analyzer := NewFrequencyAnalyzer(1.0)

	require.NoError(t, registry.Register(analyzer, 1.0))
	assert.Equal(t, 1, registry.Count())

	err := registry.Unregister("frequency_analysis")
	require.NoError(t, err)
	assert.Equal(t, 0, registry.Count())

	// Test unregistering non-existent algorithm
	err = registry.Unregister("nonexistent")
	assert.Error(t, err)
}

func TestRegistry_Clear(t *testing.T) {
	registry := NewRegistry()

	analyzer1 := NewFrequencyAnalyzer(1.0)
	analyzer2 := NewHotColdAnalyzer(1.2)

	require.NoError(t, registry.Register(analyzer1, 1.0))
	require.NoError(t, registry.Register(analyzer2, 1.2))

	assert.Equal(t, 2, registry.Count())

	registry.Clear()
	assert.Equal(t, 0, registry.Count())
}

func TestRegistry_RegisterWithNegativeWeight(t *testing.T) {
	registry := NewRegistry()
	analyzer := NewFrequencyAnalyzer(-1.0)

	err := registry.Register(analyzer, -1.0)
	assert.Error(t, err)
	assert.Equal(t, 0, registry.Count())
}

func TestEnsemble_GeneratePredictions(t *testing.T) {
	registry := NewRegistry()
	analyzer1 := NewFrequencyAnalyzer(1.0)
	analyzer2 := NewHotColdAnalyzer(1.2)

	require.NoError(t, registry.Register(analyzer1, 1.0))
	require.NoError(t, registry.Register(analyzer2, 1.2))

	ensemble := NewEnsemble(registry, WeightedVoting)
	draws := createMockDraws(valueobject.Mega645, 150)

	ctx := context.Background()
	prediction, err := ensemble.GeneratePredictions(ctx, valueobject.Mega645, draws)

	require.NoError(t, err)
	assert.NotNil(t, prediction)
	assert.Equal(t, valueobject.Mega645, prediction.GameType)
	assert.Equal(t, "weighted", prediction.VotingStrategy)
	assert.Equal(t, 6, len(prediction.FinalNumbers))
	assert.GreaterOrEqual(t, len(prediction.Predictions), 2)
	assert.Equal(t, len(prediction.Predictions), len(prediction.AlgorithmStats))
}

func TestEnsemble_GeneratePredictions_EmptyRegistry(t *testing.T) {
	registry := NewRegistry()
	ensemble := NewEnsemble(registry, WeightedVoting)
	draws := createMockDraws(valueobject.Mega645, 150)

	ctx := context.Background()
	_, err := ensemble.GeneratePredictions(ctx, valueobject.Mega645, draws)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no algorithms registered")
}

func TestEnsemble_VotingStrategies(t *testing.T) {
	registry := NewRegistry()
	analyzer1 := NewFrequencyAnalyzer(1.0)
	analyzer2 := NewHotColdAnalyzer(1.2)
	analyzer3 := NewPatternAnalyzer(0.8)

	require.NoError(t, registry.Register(analyzer1, 1.0))
	require.NoError(t, registry.Register(analyzer2, 1.2))
	require.NoError(t, registry.Register(analyzer3, 0.8))

	draws := createMockDraws(valueobject.Mega645, 150)
	ctx := context.Background()

	// Test weighted voting
	ensemble := NewEnsemble(registry, WeightedVoting)
	prediction, err := ensemble.GeneratePredictions(ctx, valueobject.Mega645, draws)
	require.NoError(t, err)
	assert.Equal(t, "weighted", prediction.VotingStrategy)
	assert.Equal(t, 6, len(prediction.FinalNumbers))

	// Test majority voting
	ensemble.SetVotingStrategy(MajorityVoting)
	prediction, err = ensemble.GeneratePredictions(ctx, valueobject.Mega645, draws)
	require.NoError(t, err)
	assert.Equal(t, "majority", prediction.VotingStrategy)

	// Test confidence weighted
	ensemble.SetVotingStrategy(ConfidenceWeighted)
	prediction, err = ensemble.GeneratePredictions(ctx, valueobject.Mega645, draws)
	require.NoError(t, err)
	assert.Equal(t, "confidence_weighted", prediction.VotingStrategy)
}

func TestEnsemble_ConsensusScore(t *testing.T) {
	registry := NewRegistry()
	analyzer1 := NewFrequencyAnalyzer(1.0)
	analyzer2 := NewHotColdAnalyzer(1.2)

	require.NoError(t, registry.Register(analyzer1, 1.0))
	require.NoError(t, registry.Register(analyzer2, 1.2))

	ensemble := NewEnsemble(registry, WeightedVoting)
	draws := createMockDraws(valueobject.Mega645, 150)

	ctx := context.Background()
	prediction, err := ensemble.GeneratePredictions(ctx, valueobject.Mega645, draws)
	require.NoError(t, err)

	consensus := ensemble.GetConsensusScore(prediction.Predictions)
	assert.GreaterOrEqual(t, consensus, 0.0)
	assert.LessOrEqual(t, consensus, 1.0)
}

func TestEnsemble_GetBestAlgorithm(t *testing.T) {
	registry := NewRegistry()
	analyzer1 := NewFrequencyAnalyzer(1.0)
	analyzer2 := NewHotColdAnalyzer(1.2)

	require.NoError(t, registry.Register(analyzer1, 1.0))
	require.NoError(t, registry.Register(analyzer2, 1.2))

	ensemble := NewEnsemble(registry, WeightedVoting)
	draws := createMockDraws(valueobject.Mega645, 150)

	ctx := context.Background()
	prediction, err := ensemble.GeneratePredictions(ctx, valueobject.Mega645, draws)
	require.NoError(t, err)

	best := ensemble.GetBestAlgorithm(prediction.AlgorithmStats)
	assert.NotNil(t, best)
	assert.NotEmpty(t, best.AlgorithmName)
	assert.GreaterOrEqual(t, best.MatchCount, 0)
}

func TestEnsemble_UpdateWeights(t *testing.T) {
	registry := NewRegistry()
	analyzer1 := NewFrequencyAnalyzer(1.0)
	analyzer2 := NewHotColdAnalyzer(1.2)

	require.NoError(t, registry.Register(analyzer1, 1.0))
	require.NoError(t, registry.Register(analyzer2, 1.2))

	ensemble := NewEnsemble(registry, WeightedVoting)

	performanceStats := map[string]float64{
		"frequency_analysis": 0.8,
		"hot_cold_analysis":  1.2,
	}

	err := ensemble.UpdateWeights(performanceStats)
	require.NoError(t, err)

	// Verify weights were updated
	assert.Equal(t, 0.8, registry.GetWeight("frequency_analysis"))
	assert.Equal(t, 1.2, registry.GetWeight("hot_cold_analysis"))
}
