package main

import (
	"context"
	"fmt"
	"time"

	"github.com/tool_predict/internal/domain/entity"
	"github.com/tool_predict/internal/domain/valueobject"
	"github.com/tool_predict/internal/infrastructure/adapter/storage"
	"github.com/tool_predict/internal/infrastructure/logger"
	"github.com/tool_predict/pkg/algorithm"
	"go.uber.org/zap"
)

// Demo prediction using local sample data
func main() {
	// Initialize logger
	logger.Init("info")
	defer logger.Sync()

	ctx := context.Background()
	gameType := valueobject.Power655

	fmt.Printf("🎯 Demo Prediction for %s (using sample data)\n\n", gameType)

	// Load sample draws from storage
	startTime := time.Now()
	jsonStorage, err := storage.NewJSONStorage("./data")
	if err != nil {
		logger.Fatal("Failed to initialize storage", zap.Error(err))
	}

	// Load draws from storage
	draws, err := jsonStorage.FindLatest(ctx, gameType, 200)
	if err != nil {
		logger.Fatal("Failed to load draws", zap.Error(err))
	}

	if len(draws) == 0 {
		logger.Fatal("No draws found. Run: python3 scripts/create_sample_draws.py")
	}

	fmt.Printf("📊 Loaded %d historical draws\n\n", len(draws))

	// Initialize algorithm registry
	registry := algorithm.NewRegistry()

	// Register algorithms
	algorithms := []struct {
		name   string
		algo   algorithm.Algorithm
		weight float64
	}{
		{"frequency_analysis", algorithm.NewFrequencyAnalyzer(1.0), 1.0},
		{"hot_cold_analysis", algorithm.NewHotColdAnalyzer(1.2), 1.2},
		{"pattern_analysis", algorithm.NewPatternAnalyzer(0.8), 0.8},
	}

	for _, alg := range algorithms {
		if err := registry.Register(alg.algo, alg.weight); err != nil {
			logger.Fatal("Failed to register algorithm",
				zap.String("algorithm", alg.name),
				zap.Error(err))
		}
		fmt.Printf("✓ Registered: %s (weight: %.1f)\n", alg.name, alg.weight)
	}

	fmt.Println()

	// Initialize ensemble
	ensemble := algorithm.NewEnsemble(registry, algorithm.WeightedVoting)

	// Train algorithms with historical data
	fmt.Println("🎓 Training algorithms with historical data...")
	for _, algo := range registry.GetAll() {
		if err := algo.Train(ctx, draws); err != nil {
			logger.Warn("Training failed",
				zap.String("algorithm", algo.Name()),
				zap.Error(err))
		} else {
			fmt.Printf("  ✓ Trained: %s\n", algo.Name())
		}
	}

	fmt.Println()

	// Generate predictions from each algorithm
	fmt.Println("🔬 Algorithm Predictions:")
	predictions := make([]*entity.Prediction, 0, registry.Count())

	for _, algo := range registry.GetAll() {
		prediction, err := algo.Predict(ctx, gameType, draws)
		if err != nil {
			logger.Warn("Prediction failed",
				zap.String("algorithm", algo.Name()),
				zap.Error(err))
			continue
		}

		predictions = append(predictions, prediction)

		fmt.Printf("  • %s: ", algo.Name())
		for i, num := range prediction.Numbers {
			fmt.Printf("%02d", num)
			if i < 5 {
				fmt.Printf(" - ")
			}
		}
		fmt.Printf(" (confidence: %.1f%%)\n", prediction.Confidence*100)
	}

	fmt.Println()

	// Generate ensemble prediction
	fmt.Println("🗳️  Ensemble Prediction (Weighted Voting):")
	ensemblePred, err := ensemble.GeneratePredictions(ctx, gameType, draws)
	if err != nil {
		logger.Fatal("Ensemble prediction failed", zap.Error(err))
	}

	// Display final prediction
	fmt.Printf("\n   🎱 FINAL PREDICTION: ")
	for i, num := range ensemblePred.FinalNumbers {
		fmt.Printf("%02d", num)
		if i < 5 {
			fmt.Printf(" - ")
		}
	}
	fmt.Printf("\n   Confidence: %.1f%%\n", calculateOverallConfidence(ensemblePred)*100)
	fmt.Printf("   Algorithms Used: %d\n", len(ensemblePred.Predictions))
	fmt.Printf("   Generated: %s\n", ensemblePred.GeneratedAt.Format("2006-01-02 15:04:05"))

	fmt.Println("\n" + "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("✅ Prediction completed in %v\n", time.Since(startTime))
}

func calculateOverallConfidence(pred *entity.EnsemblePrediction) float64 {
	if len(pred.Predictions) == 0 {
		return 0.0
	}

	totalConfidence := 0.0
	for _, p := range pred.Predictions {
		totalConfidence += p.Confidence
	}
	return (totalConfidence / float64(len(pred.Predictions)))
}
