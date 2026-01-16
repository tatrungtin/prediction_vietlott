# Deployment Guide

This guide covers deploying the Vietlott SMS Prediction Tool in production using GitHub Actions CI/CD.

## Table of Contents

1. [Prerequisites](#prerequisites)
2. [Configuration](#configuration)
3. [GitHub Actions Setup](#github-actions-setup)
4. [Local Deployment](#local-deployment)
5. [Production Deployment](#production-deployment)
6. [Monitoring](#monitoring)
7. [Troubleshooting](#troubleshooting)

---

## Prerequisites

### Required Software

- **Go**: Version 1.21 or higher
- **Git**: For version control
- **Make**: For build automation
- **Protoc**: For Protocol Buffer compilation (if modifying proto files)

### External Services

- **Vietlott API/Website**: For fetching historical lottery data
- **too_predict gRPC Service**: For sending predictions (optional)
- **GitHub Account**: For GitHub Actions CI/CD

### Required GitHub Secrets

Configure these secrets in your GitHub repository settings (`Settings > Secrets and variables > Actions`):

| Secret Name | Description | Example |
|-------------|-------------|---------|
| `GRPC_SERVER_ADDRESS` | gRPC server address for too_predict | `localhost:50051` or `too_predict.example.com:50051` |

---

## Configuration

### Environment Files

The application uses YAML configuration files:

- `configs/config.dev.yaml` - Development environment
- `configs/config.prod.yaml` - Production environment

### Production Configuration

Create `configs/config.prod.yaml`:

```yaml
app:
  name: "tool_predict"
  environment: "production"
  log_level: "info"

scraper:
  vietlott:
    base_url: "https://vietlott.vn"
    timeout: 30s
    retry_count: 3
    rate_limit: 2  # Max 2 requests per second

grpc:
  too_predict:
    address: "${GRPC_SERVER_ADDRESS}"  # Set via environment variable
    timeout: 10s

storage:
  json:
    base_path: "./data"

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
  voting_strategy: "weighted"  # Options: weighted, majority, confidence_weighted

backtest:
  default_test_period_days: 30
  default_test_period_draws: 30
```

---

## GitHub Actions Setup

### 1. Enable GitHub Actions

1. Go to your repository on GitHub
2. Click on **Actions** tab
3. If prompted, click **I understand my workflows, go ahead and enable them**

### 2. Configure Secrets

1. Go to **Settings > Secrets and variables > Actions**
2. Click **New repository secret**
3. Add the following secrets:

```bash
# gRPC server address (optional)
GRPC_SERVER_ADDRESS=localhost:50051
```

### 3. Workflow Schedules

The workflows are automatically triggered:

| Workflow | Schedule | Description |
|----------|----------|-------------|
| `daily-prediction.yml` | Daily at 18:00 UTC | Generates predictions for both game types |
| `backtest.yml` | Weekly on Sundays at 00:00 UTC | Runs comprehensive backtests |
| `ci.yml` | On push/PR | Runs tests, linting, and builds |

**Note**: Vietnam timezone is UTC+7, so 18:00 UTC = 01:00 Vietnam time (next day).

### 4. Manual Workflow Triggers

You can manually trigger workflows:

#### Daily Prediction (Manual)

1. Go to **Actions** tab
2. Select **Daily Prediction** workflow
3. Click **Run workflow**
4. Choose game type: `MEGA_6_45`, `POWER_6_55`, or `BOTH`
5. Click **Run workflow**

#### Weekly Backtest (Manual)

1. Go to **Actions** tab
2. Select **Weekly Backtest** workflow
3. Click **Run workflow**
4. Configure:
   - Game type: `MEGA_6_45`, `POWER_6_55`, or `BOTH`
   - Test size: Number of draws/days (default: 30)
   - Test mode: `draws` or `days`
5. Click **Run workflow**

---

## Local Deployment

### 1. Clone Repository

```bash
git clone https://github.com/your-username/tool_predict.git
cd tool_predict
```

### 2. Install Dependencies

```bash
go mod download
```

### 3. Generate Proto Files

```bash
make proto
```

### 4. Build Applications

```bash
make build
```

This creates:
- `bin/predictor` - Prediction CLI
- `bin/backtester` - Backtesting CLI

### 5. Run Predictions

```bash
# Generate prediction for MEGA 6/45
./bin/predictor --game-type=MEGA_6_45

# Generate prediction for POWER 6/55
./bin/predictor --game-type=POWER_6_55

# Verbose output
./bin/predictor --game-type=MEGA_6_45 --verbose
```

### 6. Run Backtests

```bash
# Backtest last 30 draws
./bin/backtester --game-type=MEGA_6_45 --test-mode=draws --test-size=30

# Backtest last 30 days
./bin/backtester --game-type=MEGA_6_45 --test-mode=days --test-size=30

# Save results to file
./bin/backtester --game-type=MEGA_6_45 --output=results.json

# Test specific algorithms
./bin/backtester --game-type=MEGA_6_45 --algorithms=frequency_analysis,hot_cold_analysis
```

---

## Production Deployment

### Option 1: GitHub Actions (Recommended)

The application is designed to run primarily on GitHub Actions:

**Pros:**
- No server management required
- Automatic scaling
- Built-in logging and artifact storage
- Free for public repositories

**Cons:**
- Limited execution time (6 hours per job)
- Requires GitHub Actions enabled

**Setup:**
1. Configure secrets as described above
2. Push code to repository
3. Workflows run automatically on schedule

### Option 2: Self-Hosted Runner

For more control, use a self-hosted GitHub Actions runner:

#### Server Setup

```bash
# Install Go 1.21+
sudo apt update
sudo apt install golang-1.21

# Clone repository
git clone https://github.com/your-username/tool_predict.git
cd tool_predict

# Build applications
make build
```

#### Add Self-Hosted Runner

1. Go to repository **Settings > Actions > Runners**
2. Click **New self-hosted runner**
3. Select OS (Linux/macOS/Windows)
4. Follow installation instructions
5. Start the runner

#### Configure Workflow to Use Self-Hosted Runner

Update workflow files:

```yaml
jobs:
  predict:
    runs-on: self-hosted  # Use self-hosted runner
    steps:
      # ... steps remain the same
```

### Option 3: Docker Deployment

For containerized deployment:

#### Dockerfile

```dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN make build

FROM alpine:latest
RUN apk --no-cache add ca-certificates

WORKDIR /app
COPY --from=builder /app/bin/predictor ./predictor
COPY --from=builder /app/bin/backtester ./backtester
COPY --from=builder /app/configs ./configs

CMD ["./predictor", "--help"]
```

#### Build and Run

```bash
# Build image
docker build -t tool_predict:latest .

# Run prediction
docker run -v $(pwd)/data:/app/data \
  -e GRPC_SERVER_ADDRESS=localhost:50051 \
  tool_predict:latest \
  ./predictor --game-type=MEGA_6_45

# Run backtest
docker run -v $(pwd)/data:/app/data \
  tool_predict:latest \
  ./backtester --game-type=MEGA_6_45 --test-mode=draws --test-size=30
```

#### Docker Compose

```yaml
version: '3.8'

services:
  predictor:
    build: .
    volumes:
      - ./data:/app/data
      - ./configs:/app/configs
    environment:
      - GRPC_SERVER_ADDRESS=${GRPC_SERVER_ADDRESS}
    command: ["./predictor", "--game-type=MEGA_6_45"]
```

---

## Monitoring

### GitHub Actions Dashboard

1. Go to **Actions** tab in repository
2. View workflow runs, logs, and artifacts
3. Download artifacts:
   - Prediction logs (retention: 30 days)
   - Prediction data (retention: 90 days)
   - Backtest results (retention: 90 days)

### Log Files

Applications log to stdout and optionally to files:

```bash
# View logs in real-time
./bin/predictor --game-type=MEGA_6_45 --verbose 2>&1 | tee prediction.log

# Save backtest results to file
./bin/backtester --game-type=MEGA_6_45 --output=backtest.json
```

### Metrics to Monitor

**Prediction Metrics:**
- Success rate (predictions generated vs. attempted)
- Execution time
- Algorithm confidence scores
- gRPC transmission success rate

**Backtest Metrics:**
- Exact match rate (6/6 numbers)
- 4-number match rate (4/6 numbers)
- 3-number match rate (3/6 numbers)
- Average confidence scores
- Algorithm performance comparison

---

## Troubleshooting

### Common Issues

#### 1. Scraper Failures

**Problem:** Failed to fetch data from Vietlott

**Solutions:**
- Check internet connectivity
- Verify Vietlott website is accessible
- Increase timeout in config: `scraper.vietlott.timeout: 60s`
- Reduce rate limit: `scraper.vietlott.rate_limit: 1`

#### 2. gRPC Connection Errors

**Problem:** Failed to send prediction via gRPC

**Solutions:**
- Verify `GRPC_SERVER_ADDRESS` is correct
- Check if too_predict service is running
- Test connection: `grpcurl -plaintext localhost:50051 list`
- Check firewall rules

#### 3. Insufficient Historical Data

**Problem:** "Not enough historical data" error

**Solutions:**
- Run historical data fetch first
- Check data directory for existing draws
- Minimum 30 draws required for backtesting

#### 4. GitHub Actions Timeout

**Problem:** Workflow exceeds time limit

**Solutions:**
- Optimize algorithms for faster execution
- Reduce test size in backtest
- Use self-hosted runner for longer execution times

#### 5. Proto Files Out of Date

**Problem:** Proto compilation errors

**Solutions:**
```bash
# Regenerate proto files
make proto

# Verify generated files exist
ls -la proto/*.pb.go

# Commit changes
git add proto/*.pb.go
git commit -m "chore: regenerate proto files"
```

### Debug Mode

Enable verbose logging:

```bash
# Set log level in config
app:
  log_level: "debug"

# Or use --verbose flag
./bin/predictor --game-type=MEGA_6_45 --verbose
```

### Getting Help

1. Check application logs
2. Review GitHub Actions logs
3. Check backtest results JSON files
4. Verify configuration files
5. Test with manual CLI commands

---

## Security Considerations

### Secrets Management

- Never commit secrets to repository
- Use GitHub Secrets for sensitive data
- Rotate credentials regularly
- Use environment variables for runtime config

### API Rate Limiting

- Respect Vietlott rate limits
- Implement exponential backoff
- Cache historical data locally
- Use appropriate delays between requests

### Data Privacy

- Lottery data is public information
- Store predictions securely
- Implement access controls for gRPC endpoints
- Use TLS for gRPC connections in production

---

## Maintenance

### Regular Tasks

**Daily:**
- Monitor GitHub Actions runs
- Check prediction logs
- Verify data quality

**Weekly:**
- Review backtest results
- Compare algorithm performance
- Update algorithm weights if needed

**Monthly:**
- Archive old prediction data
- Clean up artifacts
- Review and optimize algorithms

### Updates

**Update Dependencies:**
```bash
go get -u ./...
go mod tidy
make test
```

**Update Go Version:**
```yaml
# In .github/workflows/*.yml
env:
  GO_VERSION: '1.22'  # Update to latest stable
```

---

## Appendix

### Example Crontab Schedules

```bash
# Run predictor daily at 6 PM Vietnam time (11 AM UTC)
0 11 * * * cd /path/to/tool_predict && ./bin/predictor --game-type=MEGA_6_45

# Run backtest weekly on Sunday at midnight Vietnam time (17:00 UTC Saturday)
0 17 * * 6 cd /path/to/tool_predict && ./bin/backtester --game-type=MEGA_6_45 --test-mode=draws --test-size=30
```

### Performance Benchmarks

Expected execution times:

| Operation | Time |
|-----------|------|
| Fetch 200 draws | 10-30 seconds |
| Generate prediction | 1-5 seconds |
| Run backtest (30 iterations) | 30-60 seconds |
| Complete prediction workflow | 30-60 seconds |
| Complete backtest workflow | 2-5 minutes |

### Storage Requirements

| Data Type | Size per file | Growth Rate |
|-----------|---------------|-------------|
| Draw record | ~1 KB | 365/year/game type |
| Prediction | ~2 KB | 365/year/game type |
| Backtest result | ~10 KB | Weekly |
| Total (1 year) | ~1-2 MB | - |
