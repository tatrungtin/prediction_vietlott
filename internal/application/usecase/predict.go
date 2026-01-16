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

// PredictUseCase orchestrates the prediction workflow
type PredictUseCase struct {
	drawRepo       repository.DrawRepository
	predictionRepo repository.PredictionRepository
	ensemble       *algorithm.Ensemble
	scraper        port.VietlottScraper
	grpcClient     port.PredictionService
}

// NewPredictUseCase creates a new prediction use case
func NewPredictUseCase(
	drawRepo repository.DrawRepository,
	predictionRepo repository.PredictionRepository,
	ensemble *algorithm.Ensemble,
	scraper port.VietlottScraper,
	grpcClient port.PredictionService,
) *PredictUseCase {
	return &PredictUseCase{
		drawRepo:       drawRepo,
		predictionRepo: predictionRepo,
		ensemble:       ensemble,
		scraper:        scraper,
		grpcClient:     grpcClient,
	}
}

// Execute generates and sends a prediction
func (uc *PredictUseCase) Execute(
	ctx context.Context,
	gameType valueobject.GameType,
	algorithmCount int,
) (*EnsembleResult, error) {
	startTime := time.Now()

	logger.Info("Starting prediction workflow",
		zap.String("game_type", string(gameType)),
	)

	// Step 1: Fetch latest historical data
	logger.Info("Fetching historical data")
	draws, err := uc.scraper.FetchLatestDraws(ctx, gameType, 200)
	if err != nil {
		// Fallback to local storage if scraper fails
		logger.Warn("Scraper failed, attempting to use local storage",
			zap.Error(err),
		)
		draws, err = uc.drawRepo.FindLatest(ctx, gameType, 200)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch historical data and no local data available: %w", err)
		}
		logger.Info("Using local storage data",
			zap.Int("draws_count", len(draws)),
		)
	}

	if len(draws) < algorithmCount {
		return nil, fmt.Errorf("insufficient historical data: need at least %d draws, got %d",
			algorithmCount, len(draws))
	}

	logger.Info("Historical data fetched",
		zap.Int("draws_count", len(draws)),
	)

	// Step 2: Generate predictions using ensemble
	logger.Info("Generating ensemble predictions")
	ensemblePred, err := uc.ensemble.GeneratePredictions(ctx, gameType, draws)
	if err != nil {
		return nil, fmt.Errorf("ensemble prediction failed: %w", err)
	}

	logger.Info("Ensemble prediction generated",
		zap.String("prediction_id", ensemblePred.ID),
		zap.Strings("numbers", formatNumbers(ensemblePred.FinalNumbers)),
		zap.String("voting_strategy", ensemblePred.VotingStrategy),
		zap.Int("algorithms_used", len(ensemblePred.Predictions)),
	)

	// Step 3: Save to repository
	logger.Info("Saving prediction to repository")
	if err := uc.predictionRepo.SaveEnsemble(ctx, ensemblePred); err != nil {
		logger.Warn("Failed to save prediction to repository",
			zap.String("prediction_id", ensemblePred.ID),
			zap.Error(err),
		)
		// Don't fail the workflow if saving fails
	}

	// Step 4: Send via gRPC to too_predict (optional)
	if uc.grpcClient != nil {
		logger.Info("Sending prediction to too_predict via gRPC")
		if err := uc.grpcClient.SendPrediction(ctx, ensemblePred); err != nil {
			logger.Warn("Failed to send prediction via gRPC (continuing without it)",
				zap.String("prediction_id", ensemblePred.ID),
				zap.Error(err),
			)
			// Don't fail the workflow if gRPC fails
		} else {
			logger.Info("Prediction sent successfully to too_predict",
				zap.String("prediction_id", ensemblePred.ID),
			)
		}
	} else {
		logger.Info("gRPC client not configured, skipping send to too_predict")
	}

	duration := time.Since(startTime)

	logger.Info("Prediction workflow completed successfully",
		zap.String("prediction_id", ensemblePred.ID),
		zap.Duration("duration", duration),
	)

	// Return result
	return &EnsembleResult{
		Prediction:     ensemblePred,
		Duration:       duration,
		DrawsUsed:      len(draws),
		AlgorithmsUsed: len(ensemblePred.Predictions),
	}, nil
}

// EnsembleResult contains the prediction result and metadata
type EnsembleResult struct {
	Prediction     *entity.EnsemblePrediction
	Duration       time.Duration
	DrawsUsed      int
	AlgorithmsUsed int
}

func formatNumbers(numbers valueobject.Numbers) []string {
	result := make([]string, len(numbers))
	for i, n := range numbers {
		result[i] = fmt.Sprintf("%02d", n)
	}
	return result
}
