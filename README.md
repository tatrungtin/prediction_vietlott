# Vietlott SMS Prediction Tool

A production-ready Go microservice for predicting Vietlott lottery numbers (Mega 6/45 and Power 6/55) using hexagonal architecture, ensemble prediction algorithms, and automated CI/CD via GitHub Actions.

## 🎯 Features

- **Multi-Game Support**: Predicts numbers for both Mega 6/45 (select 6 from 01-45) and Power 6/55 (select 6 from 01-55)
- **Ensemble Algorithms**: Combines three prediction algorithms with configurable voting strategies
- **Automated Predictions**: Daily scheduled runs via GitHub Actions at 18:00 UTC
- **Comprehensive Backtesting**: Validates predictions against historical data (30 draws AND 30 calendar days)
- **gRPC Integration**: Sends predictions to the `too_predict` microservice
- **Hexagonal Architecture**: Clean separation of concerns with domain, application, and infrastructure layers
- **High Test Coverage**: 79% algorithm coverage with comprehensive unit and integration tests

## 📊 Project Status

✅ **Production Ready** - All phases completed

- [x] Phase 1: Foundation (domain, entities, repositories)
- [x] Phase 2: Algorithms (3 algorithms with ensemble voting)
- [x] Phase 3: Infrastructure (scrapers, gRPC, storage)
- [x] Phase 4: Use Cases & CLI (prediction and backtesting tools)
- [x] Phase 5: CI/CD & Deployment (GitHub Actions workflows)

## 🏗️ Architecture

```
tool_predict/
├── cmd/                          # Application entry points
│   ├── predictor/main.go         # Daily prediction CLI
│   └── backtester/main.go        # Backtesting CLI
│
├── internal/
│   ├── domain/                   # Core business logic (no dependencies)
│   │   ├── entity/               # Draw, Prediction, BacktestResult, AlgorithmStats
│   │   ├── valueobject/          # GameType, Numbers, DateRange
│   │   └── repository/           # Repository interfaces
│   │
│   ├── application/              # Business logic
│   │   ├── usecase/              # PredictUseCase, BacktestUseCase, FetchHistoricalDataUseCase
│   │   ├── port/                 # Service interfaces
│   │   └── dto/                  # Data transfer objects
│   │
│   └── infrastructure/           # External adapters
│       ├── adapter/
│       │   ├── scraper/          # Vietlott API + web scraper
│       │   ├── grpc/client/      # gRPC client for too_predict
│       │   └── storage/          # JSON file storage
│       ├── config/               # Viper configuration
│       └── logger/               # Zap structured logging
│
├── pkg/                          # Shared packages
│   ├── algorithm/                # Prediction algorithms
│   │   ├── frequency_analyzer.go
│   │   ├── hot_cold_analyzer.go
│   │   ├── pattern_analyzer.go
│   │   ├── ensemble.go
│   │   └── registry.go
│   └── grpc/proto/               # Protocol buffer definitions
│
├── .github/workflows/            # GitHub Actions CI/CD
│   ├── daily-prediction.yml      # Daily automated predictions
│   ├── backtest.yml              # Weekly backtesting
│   └── ci.yml                    # CI/CD pipeline
│
├── configs/                      # Configuration files
│   ├── config.dev.yaml
│   └── config.prod.yaml
│
├── docs/                         # Documentation
│   ├── ARCHITECTURE.md
│   ├── API.md
│   └── ALGORITHMS.md
│
├── Makefile                      # Build automation
├── go.mod
└── go.sum
```

## 🚀 Quick Start

### Prerequisites

- **Go**: Version 1.21 or higher
- **Make**: For build automation
- **Protoc**: For Protocol Buffer compilation (if modifying proto files)

### Installation

```bash
# Clone the repository
git clone https://github.com/your-org/tool_predict.git
cd tool_predict

# Install dependencies
go mod download

# Generate proto files
make proto

# Build the binaries
make build
```

This creates:
- `bin/predictor` - Prediction CLI (19MB)
- `bin/backtester` - Backtesting CLI (11MB)

### Configuration

Edit `configs/config.dev.yaml` or `configs/config.prod.yaml`:

```yaml
app:
  log_level: "info"

scraper:
  vietlott:
    base_url: "https://vietlott.vn"
    timeout: 30s
    retry_count: 3
    rate_limit: 2

grpc:
  too_predict:
    address: "localhost:50051"  # Set via GRPC_SERVER_ADDRESS env var

algorithms:
  enabled:
    - "frequency_analysis"
    - "hot_cold_analysis"
    - "pattern_analysis"
  frequency_analysis:
    weight: 1.0
  hot_cold_analysis:
    weight: 1.2
  pattern_analysis:
    weight: 0.8

ensemble:
  voting_strategy: "weighted"  # weighted, majority, confidence_weighted
```

### Running Locally

```bash
# Generate prediction for Mega 6/45
./bin/predictor --game-type=MEGA_6_45

# Generate prediction for Power 6/55
./bin/predictor --game-type=POWER_6_55

# Verbose output
./bin/predictor --game-type=MEGA_6_45 --verbose

# Run backtest - 30 draws
./bin/backtester --game-type=MEGA_6_45 --test-mode=draws --test-size=30

# Run backtest - 30 days
./bin/backtester --game-type=MEGA_6_45 --test-mode=days --test-size=30

# Save backtest results to JSON
./bin/backtester --game-type=MEGA_6_45 --output=results.json

# Test specific algorithms
./bin/backtester --game-type=MEGA_6_45 --algorithms=frequency_analysis,hot_cold_analysis
```

## 🧪 Development

### Running Tests

```bash
# Run all tests
make test

# Run with coverage
make test-coverage

# Run specific package tests
go test -v ./pkg/algorithm/...

# Run integration tests
go test -v -tags=integration ./...
```

### Code Quality

```bash
# Format code
make fmt

# Run linter
make lint

# Run all checks
make check

# Run go vet
go vet ./...
```

### Generating Proto Files

```bash
# Generate Go code from proto files
make proto

# Or manually:
protoc --go_out=. --go-grpc_out=. \
  --go_opt=paths=source_relative \
  --go-grpc_opt=paths=source_relative \
  proto/*.proto
```

## 🤖 Algorithms

### Implemented Algorithms

1. **Frequency Analyzer** (`pkg/algorithm/frequency_analyzer.go`)
   - Tracks number frequency in historical draws
   - Selects top 6 most frequent numbers
   - Weight: 1.0 (default)

2. **Hot/Cold Analyzer** (`pkg/algorithm/hot_cold_analyzer.go`)
   - Identifies recently drawn (hot) vs overdue (cold) numbers
   - Combines 3 hot + 3 cold numbers
   - Weight: 1.2 (default)

3. **Pattern Analyzer** (`pkg/algorithm/pattern_analyzer.go`)
   - Analyzes consecutive numbers, odd/even ratios, sum ranges
   - Combines multiple patterns for prediction
   - Weight: 0.8 (default)

### Ensemble Voting Strategies

- **Weighted Voting**: Uses algorithm weights for vote calculation
- **Majority Voting**: Most common numbers across all algorithms
- **Confidence Weighted**: Weights votes by algorithm confidence scores

### Algorithm Performance

See `docs/ALGORITHMS.md` for detailed algorithm descriptions and performance metrics.

## 🚢 Deployment

### GitHub Actions (Recommended)

The application runs automatically on GitHub Actions:

| Workflow | Schedule | Description |
|----------|----------|-------------|
| [Daily Prediction](.github/workflows/daily-prediction.yml) | Daily at 18:00 UTC | Generates predictions for both game types |
| [Weekly Backtest](.github/workflows/backtest.yml) | Sundays at 00:00 UTC | Comprehensive backtesting (30 draws + 30 days) |
| [CI/CD](.github/workflows/ci.yml) | On push/PR | Tests, linting, security scanning |

#### Setup GitHub Actions

1. **Configure Secrets** (Settings > Secrets and variables > Actions):
   ```bash
   GRPC_SERVER_ADDRESS=localhost:50051  # Optional, for too_predict integration
   ```

2. **Enable Workflows**: Push to repository, workflows run automatically

3. **Manual Trigger**: Go to Actions tab > Select workflow > Run workflow

### Self-Hosted Deployment

For more control, deploy on your own server:

```bash
# Clone repository
git clone https://github.com/your-org/tool_predict.git
cd tool_predict

# Build applications
make build

# Set up cron jobs
crontab -e

# Add daily prediction (6 PM Vietnam time = 11 AM UTC)
0 11 * * * cd /path/to/tool_predict && ./bin/predictor --game-type=MEGA_6_45

# Add weekly backtest (Sunday midnight Vietnam = 17:00 UTC Saturday)
0 17 * * 6 cd /path/to/tool_predict && ./bin/backtester --game-type=MEGA_6_45 --test-mode=draws --test-size=30
```

### Docker Deployment

```bash
# Build Docker image
docker build -t tool_predict:latest .

# Run prediction
docker run -v $(pwd)/data:/app/data \
  tool_predict:latest \
  ./predictor --game-type=MEGA_6_45

# Run backtest
docker run -v $(pwd)/data:/app/data \
  tool_predict:latest \
  ./backtester --game-type=MEGA_6_45 --test-mode=draws --test-size=30
```

## 📈 Monitoring

### GitHub Actions Dashboard

- View workflow runs, logs, and artifacts in Actions tab
- Download prediction logs (retention: 30 days)
- Download backtest results (retention: 90 days)

### Key Metrics

**Prediction Metrics:**
- Success rate
- Execution time
- Algorithm confidence scores
- gRPC transmission rate

**Backtest Metrics:**
- Exact match rate (6/6)
- 4-number match rate (4/6)
- 3-number match rate (3/6)
- Average confidence

## 📚 Documentation

- [Architecture](docs/ARCHITECTURE.md) - System architecture overview
- [Deployment Guide](DEPLOYMENT.md) - Complete deployment instructions
- [API Documentation](docs/API.md) - gRPC API reference
- [Algorithms](docs/ALGORITHMS.md) - Algorithm descriptions

## 🔧 Troubleshooting

### Common Issues

1. **Scraper failures**: Check internet connection, increase timeout
2. **gRPC errors**: Verify `GRPC_SERVER_ADDRESS`, check too_predict service
3. **Insufficient data**: Need at least 30 historical draws for backtesting
4. **Proto compilation**: Run `make proto` to regenerate proto files

### Debug Mode

```bash
# Enable verbose logging
./bin/predictor --game-type=MEGA_6_45 --verbose

# Or set in config
app:
  log_level: "debug"
```

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Development Guidelines

- Follow Go best practices and idioms
- Write tests for new functionality (>80% coverage target)
- Run `make check` before committing
- Update documentation for API changes
- Use conventional commit messages

## 📊 Current Status

- **Version**: 1.0.0
- **Go Version**: 1.21
- **Test Coverage**: 79% (algorithm package)
- **Last Updated**: January 2026

## ⚠️ Disclaimer

This tool is for educational and entertainment purposes only. Lottery prediction is probabilistic and past results do not guarantee future performance. Please gamble responsibly.

## 📄 License

[Your License Here]

## 🙏 Acknowledgments

- Built with [Hexagonal Architecture](https://alistair.cockburn.us/hexagonal-architecture/)
- Uses [Cobra](https://github.com/spf13/cobra) for CLI
- Logging with [Uber Zap](https://github.com/uber-go/zap)
- Configuration with [Viper](https://github.com/spf13/viper)
- gRPC with [Protocol Buffers](https://grpc.io/)
