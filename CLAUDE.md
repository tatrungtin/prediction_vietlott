# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**tool_predict** is a Go microservice for predicting Vietlott lottery numbers (Mega 6/45 and Power 6/55). It uses hexagonal architecture with ensemble prediction algorithms (frequency, hot/cold, and pattern analysis) and automated CI/CD via GitHub Actions.

## Build and Development Commands

```bash
# Build both binaries (predictor and backtester)
make build

# Run all tests with coverage
make test
make test-coverage           # Generates coverage.html

# Run specific package tests
go test -v ./pkg/algorithm/...
go test -v ./internal/application/usecase/...

# Run integration tests
go test -v -tags=integration ./...

# Code quality checks
make fmt                     # Format code with gofmt
make lint                    # Run golangci-lint
make vet                     # Run go vet
make check                   # Run all checks (fmt, vet, lint, test)

# Generate protobuf files
make proto

# Clean build artifacts
make clean
```

## Running the Applications

```bash
# Prediction CLI
./bin/predictor --game-type=MEGA_6_45       # Mega 6/45 (01-45)
./bin/predictor --game-type=POWER_6_55      # Power 6/55 (01-55)
./bin/predictor --game-type=MEGA_6_45 --verbose

# Backtesting CLI
./bin/backtester --game-type=MEGA_6_45 --test-mode=draws --test-size=30
./bin/backtester --game-type=MEGA_6_45 --test-mode=days --test-size=30
./bin/backtester --game-type=MEGA_6_45 --output=results.json
./bin/backtester --game-type=MEGA_6_45 --algorithms=frequency_analysis,hot_cold_analysis
```

## Architecture: Hexagonal/Clean Architecture

The codebase follows hexagonal architecture with strict separation of concerns:

### Three Layers

1. **Domain Layer** (`internal/domain/`)
   - Core business logic with NO external dependencies
   - **Entities**: Draw, Prediction, BacktestResult, AlgorithmStats
   - **Value Objects**: GameType, Numbers, DateRange (immutable types)
   - **Repository Interfaces**: Abstract contracts for data access

2. **Application Layer** (`internal/application/`)
   - Business logic orchestrators
   - **Use Cases**: PredictUseCase, BacktestUseCase, FetchHistoricalDataUseCase
   - **Ports**: Service interfaces for dependency inversion (VietlottScraper, Storage, PredictionService)
   - **DTOs**: Data transfer objects for layer boundaries

3. **Infrastructure Layer** (`internal/infrastructure/`)
   - External adapters implementing the ports
   - **Scrapers**: VietlottAPIScraper, VietlottWebScraper (fallback)
   - **Storage**: JSON file-based storage for draws, predictions, stats, backtests
   - **gRPC**: TooPredictClient for microservice communication
   - **Config**: Viper-based configuration (reads from `configs/config.{dev,prod}.yaml`)
   - **Logger**: Uber Zap structured logging

### Key Pattern: Dependency Inversion

The application layer depends on interfaces (ports), not concrete implementations. Infrastructure implements these interfaces. This makes the code testable and flexible.

Example:
- `PredictUseCase` depends on `DrawRepository` interface (domain layer)
- `JSONStorage` implements `DrawRepository` (infrastructure layer)

## Prediction Algorithms (`pkg/algorithm/`)

Three algorithms with configurable weights (in config YAML):

1. **Frequency Analyzer** (weight: 1.0)
   - Tracks number frequency in historical draws
   - Selects top 6 most frequent numbers

2. **Hot/Cold Analyzer** (weight: 1.2)
   - Identifies recently drawn (hot) vs overdue (cold) numbers
   - Combines 3 hot + 3 cold numbers

3. **Pattern Analyzer** (weight: 0.8)
   - Analyzes consecutive numbers, odd/even ratios, sum ranges
   - Combines multiple patterns

Ensemble combines these using voting strategies:
- **weighted**: Uses algorithm weights
- **majority**: Most common numbers across algorithms
- **confidence_weighted**: Weights by algorithm confidence scores

## Configuration

Configuration is loaded via Viper from `configs/config.{dev,prod}.yaml`. Key settings:

- **scraper.vietlott**: Timeout, retry count, rate limiting
- **grpc.too_predict.address**: Server address (override via `GRPC_SERVER_ADDRESS` env var)
- **algorithms**: Enable/disable and set weights for each algorithm
- **ensemble.voting_strategy**: How to combine algorithm outputs

## Data Storage

All data persists as JSON files in the `data/` directory:
- `data/draws/`: Historical draw results
- `data/predictions/`: Generated predictions
- `data/stats/`: Algorithm performance statistics
- `data/backtests/`: Backtest results

Minimum 30 historical draws required for reliable predictions (8 for frequency analyzer).

## Game Types

- `MEGA_6_45`: Select 6 numbers from 01-45
- `POWER_6_55`: Select 6 numbers from 01-55

## CI/CD Workflows

GitHub Actions (`.github/workflows/`):
- **daily-prediction.yml**: Runs at 18:00 UTC, generates predictions for both game types
- **backtest.yml**: Runs Sundays at 00:00 UTC, validates algorithm performance
- **ci.yml**: On push/PR, runs tests, linting, and build verification

## Important Notes

- **Minimum data requirement**: 30 historical draws for predictions, 8 for frequency analyzer
- **gRPC integration is optional**: Application can run standalone without the `too_predict` microservice
- **Scraper fallback**: If API fails, web scraper automatically attempts to fetch data
- **Lottery prediction is probabilistic**: This is educational/entertainment, past results don't guarantee future performance

## Testing Strategy

- Unit tests for individual algorithms (79% coverage in pkg/algorithm/)
- Integration tests for use cases and infrastructure
- Use `make test-coverage` to generate coverage.html
- Always run `make check` before committing

## Common Tasks

- **Add a new algorithm**: Create in `pkg/algorithm/`, register in `registry.go`, add weight to config
- **Modify voting strategy**: Update `ensemble.go` and config YAML
- **Add new game type**: Update `GameType` value object, adjust number ranges in algorithms
- **Debug scraper**: Run with `--verbose` flag or set `app.log_level: debug` in config