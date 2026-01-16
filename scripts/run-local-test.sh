#!/bin/bash

# Local Test Script - Runs predictions and backtests locally
# Useful for testing before deploying to GitHub Actions

set -e

echo "🧪 Local Testing Script"
echo "======================="
echo ""

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

GAME_TYPE=${1:-MEGA_6_45}

echo -e "${BLUE}Configuration:${NC}"
echo "  Game Type: $GAME_TYPE"
echo "  Config: ./configs/config.dev.yaml"
echo ""

# Check if binaries exist
if [ ! -f "bin/predictor" ]; then
    echo -e "${RED}✗ Predictor binary not found. Run 'make build' first${NC}"
    exit 1
fi

if [ ! -f "bin/backtester" ]; then
    echo -e "${RED}✗ Backtester binary not found. Run 'make build' first${NC}"
    exit 1
fi

# Create data directory
mkdir -p data results logs

echo -e "${BLUE}Test 1: Generate Prediction for $GAME_TYPE${NC}"
echo "----------------------------------------"
./bin/predictor \
    --config=./configs/config.dev.yaml \
    --game-type=$GAME_TYPE \
    2>&1 | tee logs/prediction_$(date +%Y%m%d_%H%M%S).log

if [ ${PIPESTATUS[0]} -eq 0 ]; then
    echo -e "${GREEN}✓ Prediction test passed${NC}"
else
    echo -e "${RED}✗ Prediction test failed${NC}"
    exit 1
fi
echo ""

echo -e "${BLUE}Test 2: Backtest - 30 Draws${NC}"
echo "----------------------------------------"
./bin/backtester \
    --config=./configs/config.dev.yaml \
    --game-type=$GAME_TYPE \
    --test-mode=draws \
    --test-size=30 \
    --output=results/backtest_${GAME_TYPE}_draws_$(date +%Y%m%d_%H%M%S).json

if [ ${PIPESTATUS[0]} -eq 0 ]; then
    echo -e "${GREEN}✓ Backtest (draws) test passed${NC}"
else
    echo -e "${RED}✗ Backtest (draws) test failed${NC}"
    exit 1
fi
echo ""

echo -e "${BLUE}Test 3: Backtest - 30 Days${NC}"
echo "----------------------------------------"
./bin/backtester \
    --config=./configs/config.dev.yaml \
    --game-type=$GAME_TYPE \
    --test-mode=days \
    --test-size=30 \
    --output=results/backtest_${GAME_TYPE}_days_$(date +%Y%m%d_%H%M%S).json

if [ ${PIPESTATUS[0]} -eq 0 ]; then
    echo -e "${GREEN}✓ Backtest (days) test passed${NC}"
else
    echo -e "${RED}✗ Backtest (days) test failed${NC}"
    exit 1
fi
echo ""

echo -e "${GREEN}✅ All tests passed successfully!${NC}"
echo ""
echo "📁 Generated Files:"
echo "  - Logs: logs/"
echo "  - Results: results/"
echo "  - Data: data/"
echo ""

# Display summary
echo "📊 Test Summary:"
if [ -d "data/predictions" ]; then
    echo "  Predictions: $(ls -1 data/predictions/ 2>/dev/null | wc -l | tr -d ' ') files"
fi
if [ -d "data/ensembles" ]; then
    echo "  Ensembles: $(ls -1 data/ensembles/ 2>/dev/null | wc -l | tr -d ' ') files"
fi
if [ -d "results" ]; then
    echo "  Backtest results: $(ls -1 results/ 2>/dev/null | wc -l | tr -d ' ') files"
fi
echo ""
