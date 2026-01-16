package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/tool_predict/internal/application/port"
	"github.com/tool_predict/internal/domain/entity"
	"github.com/tool_predict/internal/domain/repository"
	"github.com/tool_predict/internal/domain/valueobject"
	"github.com/tool_predict/internal/infrastructure/logger"
	"github.com/tool_predict/pkg/algorithm"
	"go.uber.org/zap"
)

// BacktestUseCase orchestrates the backtesting workflow
type BacktestUseCase struct {
	drawRepo     repository.DrawRepository
	backtestRepo repository.BacktestRepository
	statsRepo    repository.StatsRepository
	registry     *algorithm.Registry
	scraper      port.VietlottScraper
}

// NewBacktestUseCase creates a new backtest use case
func NewBacktestUseCase(
	drawRepo repository.DrawRepository,
	backtestRepo repository.BacktestRepository,
	statsRepo repository.StatsRepository,
	registry *algorithm.Registry,
	scraper port.VietlottScraper,
) *BacktestUseCase {
	return &BacktestUseCase{
		drawRepo:     drawRepo,
		backtestRepo: backtestRepo,
		statsRepo:    statsRepo,
		registry:     registry,
		scraper:      scraper,
	}
}

// BacktestRequest contains the backtest parameters
type BacktestRequest struct {
	GameType   valueobject.GameType
	TestMode   string // "draws" or "days"
	TestSize   int
	Algorithms []string
	FromDate   *time.Time
	ToDate     *time.Time
}

// BacktestResult contains the backtest results
type BacktestResult struct {
	GameType         valueobject.GameType
	TestMode         string
	TestPeriod       string
	TotalPredictions int
	Results          []*entity.BacktestResult
	Duration         time.Duration
}

// Execute runs the backtest
func (uc *BacktestUseCase) Execute(
	ctx context.Context,
	req BacktestRequest,
) (*BacktestResult, error) {
	startTime := time.Now()

	logger.Info("Starting backtest workflow",
		zap.String("game_type", string(req.GameType)),
		zap.String("test_mode", req.TestMode),
		zap.Int("test_size", req.TestSize),
	)

	// Step 1: Determine test period
	draws, testPeriodDesc, err := uc.getTestDraws(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get test draws: %w", err)
	}

	logger.Info("Test period determined",
		zap.String("period", testPeriodDesc),
		zap.Int("draws_count", len(draws)),
	)

	// Step 2: For each algorithm, run backtest
	algorithms := uc.registry.GetAll()
	results := make([]*entity.BacktestResult, 0, len(algorithms))

	for _, algo := range algorithms {
		if len(req.Algorithms) > 0 {
			// Filter if specific algorithms requested
			found := false
			for _, requested := range req.Algorithms {
				if algo.Name() == requested {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}

		logger.Info("Backtesting algorithm",
			zap.String("algorithm", algo.Name()),
		)

		result, err := uc.backtestAlgorithm(ctx, req.GameType, algo, draws)
		if err != nil {
			logger.Warn("Algorithm backtest failed",
				zap.String("algorithm", algo.Name()),
				zap.Error(err),
			)
			continue
		}

		results = append(results, result)
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("no backtest results generated")
	}

	duration := time.Since(startTime)

	logger.Info("Backtest workflow completed",
		zap.Int("algorithms_tested", len(results)),
		zap.Duration("duration", duration),
	)

	return &BacktestResult{
		GameType:         req.GameType,
		TestMode:         req.TestMode,
		TestPeriod:       testPeriodDesc,
		TotalPredictions: len(draws),
		Results:          results,
		Duration:         duration,
	}, nil
}

// getTestDraws gets the draws for the test period
func (uc *BacktestUseCase) getTestDraws(
	ctx context.Context,
	req BacktestRequest,
) ([]*entity.Draw, string, error) {
	if req.TestMode == "draws" {
		// Get last N draws
		draws, err := uc.scraper.FetchLatestDraws(ctx, req.GameType, req.TestSize)
		if err != nil {
			return nil, "", err
		}
		return draws, fmt.Sprintf("Last %d draws", req.TestSize), nil
	} else if req.TestMode == "days" {
		// Get draws within last N days
		fromDate := time.Now().AddDate(0, 0, -req.TestSize)
		toDate := time.Now()

		draws, err := uc.scraper.FetchDrawsByDateRange(ctx, req.GameType, fromDate, toDate)
		if err != nil {
			return nil, "", err
		}
		return draws, fmt.Sprintf("Last %d days", req.TestSize), nil
	} else if req.FromDate != nil && req.ToDate != nil {
		// Custom date range
		draws, err := uc.scraper.FetchDrawsByDateRange(ctx, req.GameType, *req.FromDate, *req.ToDate)
		if err != nil {
			return nil, "", err
		}
		return draws, fmt.Sprintf("%s to %s", req.FromDate.Format("2006-01-02"), req.ToDate.Format("2006-01-02")), nil
	}

	return nil, "", fmt.Errorf("invalid test mode: %s", req.TestMode)
}

// backtestAlgorithm backtests a single algorithm
func (uc *BacktestUseCase) backtestAlgorithm(
	ctx context.Context,
	gameType valueobject.GameType,
	algo algorithm.Algorithm,
	draws []*entity.Draw,
) (*entity.BacktestResult, error) {
	// Create test period range
	startDate := draws[0].DrawDate
	endDate := draws[len(draws)-1].DrawDate
	dateRange, _ := valueobject.NewDateRange(startDate, endDate)

	result, err := entity.NewBacktestResult(
		gameType,
		algo.Name(),
		dateRange,
		len(draws),
	)
	if err != nil {
		return nil, err
	}

	// Walk through each draw (except last few used for training)
	minTrainingDraws := 30
	if len(draws) <= minTrainingDraws {
		return nil, fmt.Errorf("not enough draws for backtesting")
	}

	for i := minTrainingDraws; i < len(draws); i++ {
		// Train on previous data
		trainingDraws := draws[:i]
		if err := algo.Train(ctx, trainingDraws); err != nil {
			logger.Warn("Training failed",
				zap.String("algorithm", algo.Name()),
				zap.Int("iteration", i),
				zap.Error(err),
			)
			continue
		}

		// Predict next draw
		actualDraw := draws[i]
		prediction, err := algo.Predict(ctx, gameType, trainingDraws)
		if err != nil {
			logger.Warn("Prediction failed",
				zap.String("algorithm", algo.Name()),
				zap.Int("iteration", i),
				zap.Error(err),
			)
			continue
		}

		// Calculate match count
		matchCount := actualDraw.Numbers.MatchCount(prediction.Numbers)

		// Record match
		match := entity.PredictionMatch{
			PredictedNumbers: prediction.Numbers,
			ActualNumbers:    actualDraw.Numbers,
			MatchCount:       matchCount,
			Confidence:       prediction.Confidence,
			PredictionDate:   prediction.GeneratedAt,
			ActualDrawDate:   actualDraw.DrawDate,
		}

		result.AddMatchResult(match)
	}

	// Calculate metrics
	result.CalculateMetrics()

	// Save to repository
	if err := uc.backtestRepo.Save(ctx, result); err != nil {
		logger.Warn("Failed to save backtest result",
			zap.String("algorithm", algo.Name()),
			zap.Error(err),
		)
	}

	logger.Info("Algorithm backtest completed",
		zap.String("algorithm", algo.Name()),
		zap.Int("exact_matches", result.ExactMatches),
		zap.Int("three_number_matches", result.ThreeNumberMatches),
		zap.Int("four_number_matches", result.FourNumberMatches),
		zap.Float64("avg_confidence", result.AverageConfidence),
	)

	return result, nil
}
