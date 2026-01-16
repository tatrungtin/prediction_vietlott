package port

import (
	"context"

	"github.com/tool_predict/internal/domain/repository"
)

// Storage combines all repository interfaces for convenience
type Storage interface {
	// Draw returns the draw repository
	Draw() repository.DrawRepository

	// Prediction returns the prediction repository
	Prediction() repository.PredictionRepository

	// Stats returns the stats repository
	Stats() repository.StatsRepository

	// Backtest returns the backtest repository
	Backtest() repository.BacktestRepository
}

// TransactionalStorage extends Storage with transaction support
type TransactionalStorage interface {
	Storage

	// BeginTx starts a new transaction
	BeginTx(ctx context.Context) (Transaction, error)
}

// Transaction represents a database transaction
type Transaction interface {
	// Commit commits the transaction
	Commit() error

	// Rollback rollbacks the transaction
	Rollback() error

	// Storage returns the storage interface for use within the transaction
	Storage() Storage
}
