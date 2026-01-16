package algorithm

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tool_predict/internal/domain/entity"
	"github.com/tool_predict/internal/domain/valueobject"
)

// Test helper function to create mock historical data
func createMockDraws(gameType valueobject.GameType, count int) []*entity.Draw {
	draws := make([]*entity.Draw, count)
	baseDate := time.Now().AddDate(0, 0, -count)

	minRange, maxRange := gameType.NumberRange()

	for i := 0; i < count; i++ {
		// Generate deterministic but varied numbers
		nums := make([]int, 6)
		for j := 0; j < 6; j++ {
			nums[j] = minRange + (i+j)%(maxRange-minRange+1)
		}

		numbers, err := valueobject.NewNumbers(nums)
		if err != nil {
			panic(err)
		}

		draw, err := entity.NewDraw(
			gameType,
			i+1,
			numbers,
			baseDate.Add(time.Duration(i)*24*time.Hour),
			float64(100000000+i*1000000),
			i%5,
		)
		if err != nil {
			panic(err)
		}

		draws[i] = draw
	}

	return draws
}

func TestFrequencyAnalyzer_Name(t *testing.T) {
	analyzer := NewFrequencyAnalyzer(1.0)
	assert.Equal(t, "frequency_analysis", analyzer.Name())
}

func TestFrequencyAnalyzer_GetWeight(t *testing.T) {
	analyzer := NewFrequencyAnalyzer(1.5)
	assert.Equal(t, 1.5, analyzer.GetWeight())
}

func TestFrequencyAnalyzer_SetWeight(t *testing.T) {
	analyzer := NewFrequencyAnalyzer(1.0)
	err := analyzer.SetWeight(2.0)
	require.NoError(t, err)
	assert.Equal(t, 2.0, analyzer.GetWeight())

	// Test negative weight
	err = analyzer.SetWeight(-1.0)
	assert.Error(t, err)
}

func TestFrequencyAnalyzer_Validate(t *testing.T) {
	analyzer := NewFrequencyAnalyzer(1.0)

	// Test with insufficient data (less than 8)
	draws := createMockDraws(valueobject.Mega645, 7)
	err := analyzer.Validate(draws)
	assert.Error(t, err)

	// Test with sufficient data
	draws = createMockDraws(valueobject.Mega645, 8)
	err = analyzer.Validate(draws)
	assert.NoError(t, err)
}

func TestFrequencyAnalyzer_Predict(t *testing.T) {
	analyzer := NewFrequencyAnalyzer(1.0)
	draws := createMockDraws(valueobject.Mega645, 100)

	ctx := context.Background()
	prediction, err := analyzer.Predict(ctx, valueobject.Mega645, draws)

	require.NoError(t, err)
	assert.NotNil(t, prediction)
	assert.Equal(t, valueobject.Mega645, prediction.GameType)
	assert.Equal(t, "frequency_analysis", prediction.AlgorithmName)
	assert.Equal(t, 6, len(prediction.Numbers))
	assert.Greater(t, prediction.Confidence, 0.0)
	assert.LessOrEqual(t, prediction.Confidence, 1.0)
}

func TestFrequencyAnalyzer_Predict_Power655(t *testing.T) {
	analyzer := NewFrequencyAnalyzer(1.0)
	draws := createMockDraws(valueobject.Power655, 100)

	ctx := context.Background()
	prediction, err := analyzer.Predict(ctx, valueobject.Power655, draws)

	require.NoError(t, err)
	assert.NotNil(t, prediction)
	assert.Equal(t, valueobject.Power655, prediction.GameType)
	assert.Equal(t, 6, len(prediction.Numbers))

	// Verify numbers are in valid range
	for _, num := range prediction.Numbers {
		assert.GreaterOrEqual(t, num, 1)
		assert.LessOrEqual(t, num, 55)
	}
}

func TestHotColdAnalyzer_Name(t *testing.T) {
	analyzer := NewHotColdAnalyzer(1.0)
	assert.Equal(t, "hot_cold_analysis", analyzer.Name())
}

func TestHotColdAnalyzer_Validate(t *testing.T) {
	analyzer := NewHotColdAnalyzer(1.0)

	// Test with insufficient data
	draws := createMockDraws(valueobject.Mega645, 30)
	err := analyzer.Validate(draws)
	assert.Error(t, err)

	// Test with sufficient data
	draws = createMockDraws(valueobject.Mega645, 60)
	err = analyzer.Validate(draws)
	assert.NoError(t, err)
}

func TestHotColdAnalyzer_Predict(t *testing.T) {
	analyzer := NewHotColdAnalyzer(1.0)
	draws := createMockDraws(valueobject.Mega645, 100)

	ctx := context.Background()
	prediction, err := analyzer.Predict(ctx, valueobject.Mega645, draws)

	require.NoError(t, err)
	assert.NotNil(t, prediction)
	assert.Equal(t, valueobject.Mega645, prediction.GameType)
	assert.Equal(t, "hot_cold_analysis", prediction.AlgorithmName)
	assert.Equal(t, 6, len(prediction.Numbers))
	assert.Greater(t, prediction.Confidence, 0.0)
}

func TestHotColdAnalyzer_Thresholds(t *testing.T) {
	analyzer := NewHotColdAnalyzer(1.0)

	// Test get/set thresholds
	assert.Equal(t, 20, analyzer.GetHotThreshold())
	assert.Equal(t, 15, analyzer.GetColdThreshold())

	err := analyzer.SetHotThreshold(25)
	require.NoError(t, err)
	assert.Equal(t, 25, analyzer.GetHotThreshold())

	err = analyzer.SetColdThreshold(20)
	require.NoError(t, err)
	assert.Equal(t, 20, analyzer.GetColdThreshold())

	// Test invalid thresholds
	err = analyzer.SetHotThreshold(3)
	assert.Error(t, err)

	err = analyzer.SetColdThreshold(2)
	assert.Error(t, err)
}

func TestPatternAnalyzer_Name(t *testing.T) {
	analyzer := NewPatternAnalyzer(1.0)
	assert.Equal(t, "pattern_analysis", analyzer.Name())
}

func TestPatternAnalyzer_Validate(t *testing.T) {
	analyzer := NewPatternAnalyzer(1.0)

	// Test with insufficient data
	draws := createMockDraws(valueobject.Mega645, 50)
	err := analyzer.Validate(draws)
	assert.Error(t, err)

	// Test with sufficient data
	draws = createMockDraws(valueobject.Mega645, 150)
	err = analyzer.Validate(draws)
	assert.NoError(t, err)
}

func TestPatternAnalyzer_Predict(t *testing.T) {
	analyzer := NewPatternAnalyzer(1.0)
	draws := createMockDraws(valueobject.Mega645, 150)

	ctx := context.Background()
	prediction, err := analyzer.Predict(ctx, valueobject.Mega645, draws)

	require.NoError(t, err)
	assert.NotNil(t, prediction)
	assert.Equal(t, valueobject.Mega645, prediction.GameType)
	assert.Equal(t, "pattern_analysis", prediction.AlgorithmName)
	assert.Equal(t, 6, len(prediction.Numbers))
	assert.Greater(t, prediction.Confidence, 0.0)
}

func TestPatternAnalyzer_Predict_Power655(t *testing.T) {
	analyzer := NewPatternAnalyzer(1.0)
	draws := createMockDraws(valueobject.Power655, 150)

	ctx := context.Background()
	prediction, err := analyzer.Predict(ctx, valueobject.Power655, draws)

	require.NoError(t, err)
	assert.NotNil(t, prediction)
	assert.Equal(t, valueobject.Power655, prediction.GameType)
	assert.Equal(t, 6, len(prediction.Numbers))

	// Verify numbers are in valid range
	for _, num := range prediction.Numbers {
		assert.GreaterOrEqual(t, num, 1)
		assert.LessOrEqual(t, num, 55)
	}
}
