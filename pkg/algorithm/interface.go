package algorithm

import (
	"context"

	"github.com/tool_predict/internal/domain/entity"
	"github.com/tool_predict/internal/domain/valueobject"
)

// Algorithm defines the interface for prediction algorithms
type Algorithm interface {
	// Name returns the algorithm identifier
	Name() string

	// Predict generates a prediction based on historical data
	Predict(
		ctx context.Context,
		gameType valueobject.GameType,
		historicalData []*entity.Draw,
	) (*entity.Prediction, error)

	// Train updates algorithm parameters based on new data
	Train(ctx context.Context, historicalData []*entity.Draw) error

	// Validate checks if algorithm can make predictions with the given data
	Validate(historicalData []*entity.Draw) error

	// GetWeight returns the algorithm's weight for ensemble voting
	GetWeight() float64

	// SetWeight sets the algorithm's weight for ensemble voting
	SetWeight(weight float64) error
}
