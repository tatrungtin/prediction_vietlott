package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/tool_predict/internal/application/usecase"
	"github.com/tool_predict/internal/domain/valueobject"
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
	testMode   string
	testSize   int
	algorithms []string
	outputFile string
)

var rootCmd = &cobra.Command{
	Use:   "backtester",
	Short: "Vietlott lottery backtesting tool",
	Long:  `A CLI tool that backtests prediction algorithms against historical data.`,
	Run:   runBacktest,
}

func init() {
	rootCmd.Flags().StringVarP(&cfgFile, "config", "c", "./configs/config.dev.yaml", "Config file path")
	rootCmd.Flags().StringVarP(&gameType, "game-type", "g", "MEGA_6_45", "Game type (MEGA_6_45 or POWER_6_55)")
	rootCmd.Flags().StringVarP(&testMode, "test-mode", "m", "draws", "Test mode (draws or days)")
	rootCmd.Flags().IntVarP(&testSize, "test-size", "s", 30, "Test size (number of draws or days)")
	rootCmd.Flags().StringSliceVarP(&algorithms, "algorithms", "a", []string{}, "Algorithms to test (default: all)")
	rootCmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output file path (JSON format)")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func runBacktest(cmd *cobra.Command, args []string) {
	// Load configuration
	cfg, err := config.Load(cfgFile)
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	if err := logger.Init(cfg.App.LogLevel); err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	logger.Info("Starting backtester application",
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

	backtestStorage, err := storage.NewBacktestJSONStorage(cfg.Storage.JSON.BasePath)
	if err != nil {
		logger.Fatal("Failed to initialize backtest storage", zap.Error(err))
		os.Exit(1)
	}

	statsStorage, err := storage.NewStatsJSONStorage(cfg.Storage.JSON.BasePath)
	if err != nil {
		logger.Fatal("Failed to initialize stats storage", zap.Error(err))
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

	// Register algorithms
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

	// Initialize use case
	backtestUseCase := usecase.NewBacktestUseCase(
		drawStorage,
		backtestStorage, // backtestRepo
		statsStorage,    // statsRepo
		registry,
		scraper,
	)

	// Create request
	req := usecase.BacktestRequest{
		GameType:   gt,
		TestMode:   testMode,
		TestSize:   testSize,
		Algorithms: algorithms,
	}

	// Execute backtest
	fmt.Printf("\n🔬 Running backtest for %s (%s: %d)...\n\n", gameType, testMode, testSize)

	startTime := time.Now()
	result, err := backtestUseCase.Execute(ctx, req)
	if err != nil {
		logger.Fatal("Backtest failed", zap.Error(err))
		os.Exit(1)
	}

	// Display results
	displayBacktestResults(result)

	duration := time.Since(startTime)
	fmt.Printf("\n✅ Backtest completed in %v\n", duration)

	// Save to file if requested
	if outputFile != "" {
		if err := saveResultsToFile(result, outputFile); err != nil {
			logger.Warn("Failed to save results to file", zap.Error(err))
		} else {
			fmt.Printf("📁 Results saved to: %s\n", outputFile)
		}
	}
}

func displayBacktestResults(result *usecase.BacktestResult) {
	fmt.Printf("📊 Backtest Results for %s\n", result.GameType)
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	fmt.Printf("Test Period:     %s\n", result.TestPeriod)
	fmt.Printf("Total Draws:     %d\n", result.TotalPredictions)
	fmt.Printf("Test Duration:   %v\n", result.Duration)
	fmt.Printf("\n")

	// Display per-algorithm results
	for _, res := range result.Results {
		fmt.Printf("🔬 %s\n", res.AlgorithmName)
		fmt.Printf("   Exact Matches (6/6):     %d\n", res.ExactMatches)
		fmt.Printf("   4-Number Matches (4/6):   %d\n", res.FourNumberMatches)
		fmt.Printf("   3-Number Matches (3/6):   %d\n", res.ThreeNumberMatches)
		fmt.Printf("   Average Confidence:       %.2f%%\n", res.AverageConfidence*100)

		// Calculate accuracy rates
		accuracy6 := float64(res.ExactMatches) / float64(res.TotalPredictions) * 100
		accuracy4 := float64(res.FourNumberMatches) / float64(res.TotalPredictions) * 100
		accuracy3 := float64(res.ThreeNumberMatches) / float64(res.TotalPredictions) * 100

		fmt.Printf("   Accuracy Rates:\n")
		fmt.Printf("      6/6:  %.2f%%\n", accuracy6)
		fmt.Printf("      4/6:  %.2f%%\n", accuracy4)
		fmt.Printf("      3/6:  %.2f%%\n", accuracy3)
		fmt.Printf("\n")
	}
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
}

func saveResultsToFile(result *usecase.BacktestResult, filename string) error {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filename, data, 0644)
}
