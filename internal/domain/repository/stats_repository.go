package repository

import (
	"context"

	"github.com/tool_predict/internal/domain/entity"
	"github.com/tool_predict/internal/domain/valueobject"
)

// StatsRepository defines the interface for algorithm statistics persistence
type StatsRepository interface {
	// Save saves algorithm statistics to the repository
	Save(ctx context.Context, stats *entity.AlgorithmStats) error

	// Find finds statistics for a specific algorithm and game type
	Find(
		ctx context.Context,
		algorithmName string,
		gameType valueobject.GameType,
	) (*entity.AlgorithmStats, error)

	// FindAll finds all algorithm statistics
	FindAll(ctx context.Context) ([]*entity.AlgorithmStats, error)

	// FindByGameType finds all statistics for a specific game type
	FindByGameType(ctx context.Context, gameType valueobject.GameType) ([]*entity.AlgorithmStats, error)

	// FindActive finds all active algorithm statistics
	FindActive(ctx context.Context) ([]*entity.AlgorithmStats, error)

	// Update updates algorithm statistics in the repository
	Update(ctx context.Context, stats *entity.AlgorithmStats) error

	// UpdateWeight updates only the weight for an algorithm
	UpdateWeight(
		ctx context.Context,
		algorithmName string,
		gameType valueobject.GameType,
		weight float64,
	) error

	// UpdateMetrics updates the performance metrics for an algorithm
	UpdateMetrics(
		ctx context.Context,
		algorithmName string,
		gameType valueobject.GameType,
		accuracy3Num float64,
		accuracy4Num float64,
		accuracyExact float64,
		avgConfidence float64,
		totalPredictions int,
	) error

	// SetActive sets the active status for an algorithm
	SetActive(
		ctx context.Context,
		algorithmName string,
		gameType valueobject.GameType,
		isActive bool,
	) error
}
