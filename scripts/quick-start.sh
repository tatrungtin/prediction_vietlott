#!/bin/bash

# Quick Start Script for tool_predict
# This script performs initial setup and runs a test prediction

set -e

echo "🎯 Vietlott SMS Prediction Tool - Quick Start"
echo "=============================================="
echo ""

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Step 1: Install dependencies
echo -e "${BLUE}Step 1: Installing dependencies...${NC}"
go mod download
echo -e "${GREEN}✓ Dependencies installed${NC}"
echo ""

# Step 2: Generate proto files
echo -e "${BLUE}Step 2: Generating proto files...${NC}"
if [ -f "Makefile" ]; then
    make proto
else
    protoc --go_out=. --go-grpc_out=. \
        --go_opt=paths=source_relative \
        --go-grpc_opt=paths=source_relative \
        proto/*.proto
fi
echo -e "${GREEN}✓ Proto files generated${NC}"
echo ""

# Step 3: Build applications
echo -e "${BLUE}Step 3: Building applications...${NC}"
mkdir -p bin
go build -o bin/predictor ./cmd/predictor
go build -o bin/backtester ./cmd/backtester
chmod +x bin/predictor bin/backtester
echo -e "${GREEN}✓ Applications built${NC}"
echo ""

# Step 4: Create data directory
echo -e "${BLUE}Step 4: Creating data directory...${NC}"
mkdir -p data/{draws,predictions,ensembles,backtests,stats}
mkdir -p data/{mega_6_45,power_6_55}
echo -e "${GREEN}✓ Data directories created${NC}"
echo ""

# Step 5: Run test prediction
echo -e "${BLUE}Step 5: Running test prediction for MEGA 6/45...${NC}"
echo ""
./bin/predictor --game-type=MEGA_6_45
echo ""
echo -e "${GREEN}✓ Test prediction completed${NC}"
echo ""

# Step 6: Display results
echo -e "${BLUE}Step 6: Prediction results:${NC}"
echo ""
if [ -d "data/predictions" ]; then
    echo "📁 Prediction files:"
    ls -lh data/predictions/ 2>/dev/null || echo "  No predictions found"
fi

if [ -d "data/ensembles" ]; then
    echo ""
    echo "📁 Ensemble files:"
    ls -lh data/ensembles/ 2>/dev/null || echo "  No ensembles found"
fi
echo ""

echo -e "${GREEN}✅ Quick start completed successfully!${NC}"
echo ""
echo "📖 Next Steps:"
echo "  - Run prediction: ./bin/predictor --game-type=MEGA_6_45"
echo "  - Run backtest: ./bin/backtester --game-type=MEGA_6_45 --test-mode=draws --test-size=30"
echo "  - View help: ./bin/predictor --help"
echo ""
echo "📚 Documentation:"
echo "  - README.md: Complete project overview"
echo "  - DEPLOYMENT.md: Deployment guide"
echo ""
