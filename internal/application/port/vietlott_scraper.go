package port

import (
	"context"
	"time"

	"github.com/tool_predict/internal/domain/entity"
	"github.com/tool_predict/internal/domain/valueobject"
)

// VietlottScraper defines the interface for scraping Vietlott lottery data
type VietlottScraper interface {
	// FetchLatestDraws fetches the most recent draws for a game type
	FetchLatestDraws(
		ctx context.Context,
		gameType valueobject.GameType,
		limit int,
	) ([]*entity.Draw, error)

	// FetchAllDraws fetches all draws from a specified date onwards
	FetchAllDraws(
		ctx context.Context,
		gameType valueobject.GameType,
		fromDate time.Time,
	) ([]*entity.Draw, error)

	// FetchDrawByNumber fetches a specific draw by its draw number
	FetchDrawByNumber(
		ctx context.Context,
		gameType valueobject.GameType,
		drawNumber int,
	) (*entity.Draw, error)

	// FetchDrawsByDateRange fetches all draws within a date range
	FetchDrawsByDateRange(
		ctx context.Context,
		gameType valueobject.GameType,
		startDate time.Time,
		endDate time.Time,
	) ([]*entity.Draw, error)

	// GetLatestDrawNumber returns the most recent draw number for a game type
	GetLatestDrawNumber(
		ctx context.Context,
		gameType valueobject.GameType,
	) (int, error)
}
