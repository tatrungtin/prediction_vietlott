package repository

import (
	"context"

	"github.com/tool_predict/internal/domain/entity"
	"github.com/tool_predict/internal/domain/valueobject"
)

// DrawRepository defines the interface for draw data persistence
type DrawRepository interface {
	// Save saves a draw to the repository
	Save(ctx context.Context, draw *entity.Draw) error

	// SaveBatch saves multiple draws in a single transaction
	SaveBatch(ctx context.Context, draws []*entity.Draw) error

	// FindByID finds a draw by its unique identifier
	FindByID(ctx context.Context, id string) (*entity.Draw, error)

	// FindByGameTypeAndDrawNumber finds a draw by game type and draw number
	FindByGameTypeAndDrawNumber(
		ctx context.Context,
		gameType valueobject.GameType,
		drawNumber int,
	) (*entity.Draw, error)

	// FindLatest finds the most recent draws for a game type
	FindLatest(
		ctx context.Context,
		gameType valueobject.GameType,
		limit int,
	) ([]*entity.Draw, error)

	// FindByDateRange finds all draws within a date range for a game type
	FindByDateRange(
		ctx context.Context,
		gameType valueobject.GameType,
		dateRange valueobject.DateRange,
	) ([]*entity.Draw, error)

	// FindByDrawNumberRange finds draws within a draw number range
	FindByDrawNumberRange(
		ctx context.Context,
		gameType valueobject.GameType,
		startDrawNumber int,
		endDrawNumber int,
	) ([]*entity.Draw, error)

	// Count returns the total number of draws for a game type
	Count(ctx context.Context, gameType valueobject.GameType) (int64, error)

	// DeleteAll deletes all draws for a game type (useful for testing)
	DeleteAll(ctx context.Context, gameType valueobject.GameType) error

	// GetLatestDrawNumber returns the highest draw number for a game type
	GetLatestDrawNumber(ctx context.Context, gameType valueobject.GameType) (int, error)
}
