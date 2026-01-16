# 🎯 Quick Reference Card

## 🚀 Common Commands

### Build & Run
```bash
# Build all binaries
make build

# Run prediction
./bin/predictor --game-type=MEGA_6_45

# Run backtest
./bin/backtester --game-type=MEGA_6_45 --test-mode=draws --test-size=30

# Quick start (build + test)
./scripts/quick-start.sh

# Local testing
./scripts/run-local-test.sh MEGA_6_45
```

### Git Commands
```bash
# Check status
git status

# View commit history
git log --oneline

# Create new branch
git checkout -b feature/your-feature

# Commit changes
git add .
git commit -m "Your message"

# Push to GitHub
git push origin main
```

### GitHub Actions Setup
```bash
# Automated setup (recommended)
./scripts/setup-github.sh

# Manual setup
git remote add origin git@github.com:USERNAME/tool_predict.git
git push -u origin main
```

## 📋 CLI Flags

### Predictor (`./bin/predictor`)
| Flag | Description | Default |
|------|-------------|---------|
| `--config` | Config file path | `./configs/config.dev.yaml` |
| `--game-type` | Game type | `MEGA_6_45` |
| `--verbose` | Verbose output | `false` |
| `--help` | Show help | - |

### Backtester (`./bin/backtester`)
| Flag | Description | Default |
|------|-------------|---------|
| `--config` | Config file path | `./configs/config.dev.yaml` |
| `--game-type` | Game type | `MEGA_6_45` |
| `--test-mode` | Test mode (draws/days) | `draws` |
| `--test-size` | Number of draws/days | `30` |
| `--algorithms` | Specific algorithms | `all` |
| `--output` | Output JSON file | - |
| `--help` | Show help | - |

## 🎮 Game Types

| Type | Range | Numbers |
|------|-------|---------|
| `MEGA_6_45` | 01-45 | Select 6 |
| `POWER_6_55` | 01-55 | Select 6 |

## 🤖 Algorithms

| Algorithm | Weight | Description |
|-----------|--------|-------------|
| `frequency_analysis` | 1.0 | Most frequent numbers |
| `hot_cold_analysis` | 1.2 | Hot + cold numbers |
| `pattern_analysis` | 0.8 | Pattern-based |

## 🗳️ Voting Strategies

| Strategy | Description |
|----------|-------------|
| `weighted` | Weighted by algorithm weights |
| `majority` | Most common across algorithms |
| `confidence_weighted` | Weighted by confidence scores |

## 📅 Workflow Schedule

| Workflow | Schedule | Description |
|----------|----------|-------------|
| Daily Prediction | 18:00 UTC | Generate predictions |
| Weekly Backtest | Sun 00:00 UTC | Run backtests |
| CI/CD | On push/PR | Test & build |

## 📁 Important Files

| File | Purpose |
|------|---------|
| `README.md` | Project overview |
| `DEPLOYMENT.md` | Deployment guide |
| `NEXT_STEPS.md` | Setup instructions |
| `configs/config.dev.yaml` | Dev config |
| `configs/config.prod.yaml` | Production config |
| `.github/workflows/` | CI/CD workflows |
| `Makefile` | Build automation |

## 🔧 Troubleshooting

### Build Errors
```bash
# Clean and rebuild
make clean
make build

# Update dependencies
go mod tidy
go mod download
```

### Proto Files
```bash
# Regenerate proto files
make proto

# Or manually
protoc --go_out=. --go-grpc_out=. \
  --go_opt=paths=source_relative \
  --go-grpc_opt=paths=source_relative \
  proto/*.proto
```

### Test Failures
```bash
# Run specific test
go test -v ./pkg/algorithm/...

# Run with race detection
go test -race ./...

# Run coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## 🌐 URLs

- **GitHub**: Create at https://github.com/new
- **Actions**: Repository > Actions tab
- **Secrets**: Settings > Secrets > Actions
- **Docs**: https://docs.github.com/en/actions

## 📊 File Sizes

- `bin/predictor`: ~19MB
- `bin/backtester`: ~11MB

## 🎯 Quick Test

```bash
# Run complete local test
./scripts/run-local-test.sh MEGA_6_45

# Expected output:
# ✓ Prediction test passed
# ✓ Backtest (draws) test passed
# ✓ Backtest (days) test passed
```

## 💡 Tips

1. **Always test locally** before GitHub deployment
2. **Check logs** in Actions tab for debugging
3. **Download artifacts** to verify results
4. **Tune weights** based on backtest results
5. **Monitor execution times** for performance

## 🆘 Getting Help

- Check logs: `./bin/predictor --verbose`
- Read docs: `README.md`, `DEPLOYMENT.md`
- Review workflows: `.github/workflows/`
- Check issues: GitHub Issues tab

---
*Last updated: January 2026*
