package repository

import (
	"context"

	"github.com/tool_predict/internal/domain/entity"
	"github.com/tool_predict/internal/domain/valueobject"
)

// PredictionRepository defines the interface for prediction data persistence
type PredictionRepository interface {
	// Save saves a single prediction to the repository
	Save(ctx context.Context, prediction *entity.Prediction) error

	// SaveBatch saves multiple predictions in a single transaction
	SaveBatch(ctx context.Context, predictions []*entity.Prediction) error

	// SaveEnsemble saves an ensemble prediction to the repository
	SaveEnsemble(ctx context.Context, ensemble *entity.EnsemblePrediction) error

	// FindByID finds a prediction by its unique identifier
	FindByID(ctx context.Context, id string) (*entity.Prediction, error)

	// FindEnsembleByID finds an ensemble prediction by its unique identifier
	FindEnsembleByID(ctx context.Context, id string) (*entity.EnsemblePrediction, error)

	// FindLatest finds the most recent predictions for a game type
	FindLatest(
		ctx context.Context,
		gameType valueobject.GameType,
		limit int,
	) ([]*entity.Prediction, error)

	// FindLatestEnsembles finds the most recent ensemble predictions for a game type
	FindLatestEnsembles(
		ctx context.Context,
		gameType valueobject.GameType,
		limit int,
	) ([]*entity.EnsemblePrediction, error)

	// FindByAlgorithm finds predictions for a specific algorithm and game type
	FindByAlgorithm(
		ctx context.Context,
		algorithmName string,
		gameType valueobject.GameType,
		limit int,
	) ([]*entity.Prediction, error)

	// FindByDateRange finds predictions within a date range for a game type
	FindByDateRange(
		ctx context.Context,
		gameType valueobject.GameType,
		startDate interface{}, // time.Time
		endDate interface{}, // time.Time
	) ([]*entity.Prediction, error)

	// Count returns the total number of predictions for a game type
	Count(ctx context.Context, gameType valueobject.GameType) (int64, error)

	// DeleteOld removes predictions older than a certain date
	DeleteOld(ctx context.Context, beforeDate interface{}) error // time.Time
}
