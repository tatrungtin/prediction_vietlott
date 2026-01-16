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
	"go.uber.org/zap"
)

// FetchHistoricalDataUseCase fetches historical lottery data from Vietlott
type FetchHistoricalDataUseCase struct {
	drawRepo repository.DrawRepository
	scraper  port.VietlottScraper
}

// NewFetchHistoricalDataUseCase creates a new use case
func NewFetchHistoricalDataUseCase(
	drawRepo repository.DrawRepository,
	scraper port.VietlottScraper,
) *FetchHistoricalDataUseCase {
	return &FetchHistoricalDataUseCase{
		drawRepo: drawRepo,
		scraper:  scraper,
	}
}

// FetchLatest fetches the latest draws for a game type
func (uc *FetchHistoricalDataUseCase) FetchLatest(
	ctx context.Context,
	gameType valueobject.GameType,
	limit int,
) ([]*entity.Draw, error) {
	logger.Info("Fetching latest draws",
		zap.String("game_type", string(gameType)),
		zap.Int("limit", limit),
	)

	// Fetch from scraper
	draws, err := uc.scraper.FetchLatestDraws(ctx, gameType, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch draws from scraper: %w", err)
	}

	// Save to repository
	for _, draw := range draws {
		if err := uc.drawRepo.Save(ctx, draw); err != nil {
			logger.Warn("Failed to save draw",
				zap.String("draw_id", draw.ID),
				zap.Error(err),
			)
			// Continue saving other draws
		}
	}

	logger.Info("Successfully fetched and saved draws",
		zap.String("game_type", string(gameType)),
		zap.Int("count", len(draws)),
	)

	return draws, nil
}

// FetchFromDate fetches all draws from a specified date onwards
func (uc *FetchHistoricalDataUseCase) FetchFromDate(
	ctx context.Context,
	gameType valueobject.GameType,
	fromDate time.Time,
) ([]*entity.Draw, error) {
	logger.Info("Fetching historical data",
		zap.String("game_type", string(gameType)),
		zap.Time("from_date", fromDate),
	)

	// Fetch from scraper
	draws, err := uc.scraper.FetchAllDraws(ctx, gameType, fromDate)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch draws from scraper: %w", err)
	}

	// Save to repository
	savedCount := 0
	for _, draw := range draws {
		if err := uc.drawRepo.Save(ctx, draw); err != nil {
			logger.Warn("Failed to save draw",
				zap.String("draw_id", draw.ID),
				zap.Error(err),
			)
			continue
		}
		savedCount++
	}

	logger.Info("Successfully fetched and saved historical draws",
		zap.String("game_type", string(gameType)),
		zap.Int("total", len(draws)),
		zap.Int("saved", savedCount),
	)

	return draws, nil
}

// FetchByRange fetches draws within a date range
func (uc *FetchHistoricalDataUseCase) FetchByRange(
	ctx context.Context,
	gameType valueobject.GameType,
	startDate time.Time,
	endDate time.Time,
) ([]*entity.Draw, error) {
	dateRange, err := valueobject.NewDateRange(startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("invalid date range: %w", err)
	}

	logger.Info("Fetching draws by date range",
		zap.String("game_type", string(gameType)),
		zap.String("range", dateRange.String()),
	)

	// Check if we already have the data in storage
	existingDraws, err := uc.drawRepo.FindByDateRange(ctx, gameType, dateRange)
	if err == nil && len(existingDraws) > 0 {
		logger.Info("Using existing data from storage",
			zap.Int("count", len(existingDraws)),
		)
		return existingDraws, nil
	}

	// Fetch from scraper
	draws, err := uc.scraper.FetchDrawsByDateRange(ctx, gameType, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch draws from scraper: %w", err)
	}

	// Save to repository
	for _, draw := range draws {
		if err := uc.drawRepo.Save(ctx, draw); err != nil {
			logger.Warn("Failed to save draw",
				zap.String("draw_id", draw.ID),
				zap.Error(err),
			)
		}
	}

	return draws, nil
}

// GetLatestDrawNumber returns the most recent draw number
func (uc *FetchHistoricalDataUseCase) GetLatestDrawNumber(
	ctx context.Context,
	gameType valueobject.GameType,
) (int, error) {
	// Try to get from storage first
	latestNum, err := uc.drawRepo.GetLatestDrawNumber(ctx, gameType)
	if err == nil && latestNum > 0 {
		return latestNum, nil
	}

	// If not in storage, fetch from scraper
	latestNum, err = uc.scraper.GetLatestDrawNumber(ctx, gameType)
	if err != nil {
		return 0, fmt.Errorf("failed to get latest draw number: %w", err)
	}

	return latestNum, nil
}
