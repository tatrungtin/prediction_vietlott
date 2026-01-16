package main

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/tool_predict/internal/application/port"
	"github.com/tool_predict/internal/application/usecase"
	"github.com/tool_predict/internal/domain/entity"
	"github.com/tool_predict/internal/domain/valueobject"
	"github.com/tool_predict/internal/infrastructure/adapter/grpc/client"
	"github.com/tool_predict/internal/infrastructure/adapter/scraper"
	"github.com/tool_predict/internal/infrastructure/adapter/storage"
	"github.com/tool_predict/internal/infrastructure/config"
	"github.com/tool_predict/internal/infrastructure/logger"
	"github.com/tool_predict/pkg/algorithm"
	"go.uber.org/zap"
)

var (
	cfgFile    string
	gameType   string
	verbose    bool
)

var rootCmd = &cobra.Command{
	Use:   "predictor",
	Short: "Vietlott lottery prediction tool",
	Long:  `A CLI tool that predicts Vietlott lottery numbers using ensemble algorithms.`,
	Run:   runPredict,
}

var predictCmd = &cobra.Command{
	Use:   "predict",
	Short: "Generate a prediction",
	Run:   runPredict,
}

func init() {
	rootCmd.Flags().StringVarP(&cfgFile, "config", "c", "./configs/config.dev.yaml", "Config file path")
	rootCmd.Flags().StringVarP(&gameType, "game-type", "g", "MEGA_6_45", "Game type (MEGA_6_45 or POWER_6_55)")
	rootCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func runPredict(cmd *cobra.Command, args []string) {
	// Load configuration
	cfg, err := config.Load(cfgFile)
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	logLevel := cfg.App.LogLevel
	if verbose {
		logLevel = "debug"
	}
	if err := logger.Init(logLevel); err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	logger.Info("Starting predictor application",
		zap.String("version", "1.0.0"),
		zap.String("environment", cfg.App.Environment),
	)

	// Parse game type
	gt := valueobject.GameType(gameType)
	if err := gt.Validate(); err != nil {
		logger.Fatal("Invalid game type", zap.Error(err))
		os.Exit(1)
	}

	// Initialize components
	ctx := context.Background()

	// Initialize storage
	drawStorage, err := storage.NewJSONStorage(cfg.Storage.JSON.BasePath)
	if err != nil {
		logger.Fatal("Failed to initialize draw storage", zap.Error(err))
		os.Exit(1)
	}

	predictionStorage, err := storage.NewPredictionJSONStorage(cfg.Storage.JSON.BasePath)
	if err != nil {
		logger.Fatal("Failed to initialize prediction storage", zap.Error(err))
		os.Exit(1)
	}

	// Initialize scraper
	scraper := scraper.NewVietlottAPIScraper(
		cfg.Scraper.Vietlott.BaseURL,
		cfg.Scraper.Vietlott.Timeout,
		cfg.Scraper.Vietlott.RetryCount,
		cfg.Scraper.Vietlott.RateLimit,
	)

	// Initialize algorithm registry
	registry := algorithm.NewRegistry()

	// Register algorithms based on config
	for _, algoName := range cfg.Algorithms.Enabled {
		var algo algorithm.Algorithm
		var weight float64

		switch algoName {
		case "frequency_analysis":
			algo = algorithm.NewFrequencyAnalyzer(
				cfg.Algorithms.Configs[algoName].Weight,
			)
			weight = cfg.Algorithms.Configs[algoName].Weight
		case "hot_cold_analysis":
			algo = algorithm.NewHotColdAnalyzer(
				cfg.Algorithms.Configs[algoName].Weight,
			)
			weight = cfg.Algorithms.Configs[algoName].Weight
		case "pattern_analysis":
			algo = algorithm.NewPatternAnalyzer(
				cfg.Algorithms.Configs[algoName].Weight,
			)
			weight = cfg.Algorithms.Configs[algoName].Weight
		default:
			logger.Warn("Unknown algorithm, skipping",
				zap.String("algorithm", algoName),
			)
			continue
		}

		if err := registry.Register(algo, weight); err != nil {
			logger.Fatal("Failed to register algorithm",
				zap.String("algorithm", algoName),
				zap.Error(err),
			)
			os.Exit(1)
		}
	}

	logger.Info("Algorithms registered",
		zap.Int("count", registry.Count()),
	)

	// Initialize ensemble
	votingStrategy := algorithm.VotingStrategy(cfg.Ensemble.VotingStrategy)
	ensemble := algorithm.NewEnsemble(registry, votingStrategy)

	// Initialize gRPC client
	var grpcClient port.PredictionService
	if cfg.GRPC.TooPredict.Address != "" {
		grpcClient, err = client.NewTooPredictClient(cfg.GRPC.TooPredict.Address)
		if err != nil {
			logger.Warn("Failed to create gRPC client, predictions will not be sent",
				zap.Error(err),
			)
			grpcClient = nil
		}
	}

	// Initialize use case
	predictUseCase := usecase.NewPredictUseCase(
		drawStorage,
		predictionStorage,
		ensemble,
		scraper,
		grpcClient,
	)

	// Execute prediction
	fmt.Printf("\n🎯 Generating prediction for %s...\n\n", gameType)

	result, err := predictUseCase.Execute(ctx, gt, registry.Count())
	if err != nil {
		logger.Fatal("Prediction failed", zap.Error(err))
		os.Exit(1)
	}

	// Display results
	displayResult(result, gt)

	fmt.Printf("\n✅ Prediction completed in %v\n", result.Duration)
}

func displayResult(result *usecase.EnsembleResult, gameType valueobject.GameType) {
	fmt.Printf("📊 Prediction Results for %s\n", gameType)
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	fmt.Printf("Prediction ID:  %s\n", result.Prediction.ID)
	fmt.Printf("Predicted Numbers:  ")
	for i, num := range result.Prediction.FinalNumbers {
		fmt.Printf("%02d", num)
		if i < 5 {
			fmt.Printf(" - ")
		}
	}
	fmt.Printf("\n")
	fmt.Printf("Voting Strategy: %s\n", result.Prediction.VotingStrategy)
	fmt.Printf("Algorithms Used:  %d\n", result.AlgorithmsUsed)
	fmt.Printf("Confidence:       %.2f%%\n", calculateOverallConfidence(result.Prediction))
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")

	// Show algorithm contributions
	fmt.Printf("\n🔬 Algorithm Contributions:\n")
	for _, stat := range result.Prediction.AlgorithmStats {
		fmt.Printf("  • %s: %d matches, confidence: %.2f%%\n",
			stat.AlgorithmName,
			stat.MatchCount,
			stat.Confidence*100,
		)
	}
}

func calculateOverallConfidence(pred *entity.EnsemblePrediction) float64 {
	if len(pred.Predictions) == 0 {
		return 0.0
	}

	totalConfidence := 0.0
	for _, p := range pred.Predictions {
		totalConfidence += p.Confidence
	}
	return (totalConfidence / float64(len(pred.Predictions))) * 100
}
