package repository

import (
	"context"

	"github.com/tool_predict/internal/domain/entity"
	"github.com/tool_predict/internal/domain/valueobject"
)

// BacktestRepository defines the interface for backtest result persistence
type BacktestRepository interface {
	// Save saves a backtest result to the repository
	Save(ctx context.Context, result *entity.BacktestResult) error

	// FindByID finds a backtest result by its unique identifier
	FindByID(ctx context.Context, id string) (*entity.BacktestResult, error)

	// FindLatest finds the most recent backtest results
	FindLatest(ctx context.Context, limit int) ([]*entity.BacktestResult, error)

	// FindByAlgorithm finds backtest results for a specific algorithm
	FindByAlgorithm(
		ctx context.Context,
		algorithmName string,
		gameType valueobject.GameType,
		limit int,
	) ([]*entity.BacktestResult, error)

	// FindByGameType finds all backtest results for a specific game type
	FindByGameType(ctx context.Context, gameType valueobject.GameType) ([]*entity.BacktestResult, error)

	// FindByDateRange finds backtest results within a date range
	FindByDateRange(
		ctx context.Context,
		startDate interface{}, // time.Time
		endDate interface{},   // time.Time
	) ([]*entity.BacktestResult, error)

	// FindBestPerforming finds the best performing algorithm for a game type
	FindBestPerforming(
		ctx context.Context,
		gameType valueobject.GameType,
		metric string, // "exact", "4_numbers", "3_numbers"
	) (*entity.BacktestResult, error)

	// DeleteOld removes backtest results older than a certain date
	DeleteOld(ctx context.Context, beforeDate interface{}) error // time.Time
}
