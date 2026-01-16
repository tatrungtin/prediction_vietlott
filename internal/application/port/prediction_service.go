package port

import (
	"context"
	"time"

	"github.com/tool_predict/internal/domain/entity"
)

// PredictionService defines the interface for sending predictions to external services
type PredictionService interface {
	// SendPrediction sends an ensemble prediction to the too_predict service
	SendPrediction(
		ctx context.Context,
		prediction *entity.EnsemblePrediction,
	) error

	// GetPredictionStatus checks the status of a sent prediction
	GetPredictionStatus(
		ctx context.Context,
		predictionID string,
	) (*PredictionStatus, error)
}

// PredictionStatus represents the status of a prediction sent to an external service
type PredictionStatus struct {
	ID          string    `json:"id"`
	Status      string    `json:"status"` // "pending", "processed", "failed"
	SentAt      time.Time `json:"sent_at"`
	ProcessedAt time.Time `json:"processed_at"`
	Error       string    `json:"error,omitempty"`
}
