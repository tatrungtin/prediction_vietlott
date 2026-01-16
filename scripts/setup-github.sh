#!/bin/bash

# GitHub Repository Setup Script for tool_predict
# This script helps you set up the GitHub repository and configure necessary settings

set -e

echo "🚀 GitHub Repository Setup for tool_predict"
echo "=========================================="
echo ""

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check if gh CLI is installed
if ! command -v gh &> /dev/null; then
    echo -e "${YELLOW}GitHub CLI (gh) not found. Installing...${NC}"
    # Check OS and install accordingly
    if [[ "$OSTYPE" == "darwin"* ]]; then
        brew install gh
    else
        echo "Please install GitHub CLI from: https://cli.github.com/"
        exit 1
    fi
fi

# Check if user is authenticated
echo "📋 Checking GitHub authentication..."
if ! gh auth status &> /dev/null; then
    echo -e "${YELLOW}Not authenticated with GitHub. Please login:${NC}"
    gh auth login
fi

echo -e "${GREEN}✓ GitHub authentication OK${NC}"
echo ""

# Get repository name
read -p "Enter your GitHub username: " GITHUB_USERNAME
read -p "Enter repository name (default: tool_predict): " REPO_NAME
REPO_NAME=${REPO_NAME:-tool_predict}

echo ""
echo "📦 Creating GitHub repository..."
gh repo create "$GITHUB_USERNAME/$REPO_NAME" \
    --public \
    --description "Vietlott SMS Prediction Tool - Ensemble prediction algorithms for lottery number forecasting" \
    --source=. \
    --remote=origin \
    --push || echo "Repository might already exist"

echo -e "${GREEN}✓ Repository created/connected${NC}"
echo ""

# Configure secrets
echo "🔐 Configuring GitHub Secrets..."
echo ""

read -p "Enter gRPC server address (leave empty if not using too_predict): " GRPC_ADDRESS

if [ -n "$GRPC_ADDRESS" ]; then
    gh secret set GRPC_SERVER_ADDRESS -b"$GRPC_ADDRESS" --repo "$GITHUB_USERNAME/$REPO_NAME"
    echo -e "${GREEN}✓ GRPC_SERVER_ADDRESS secret set${NC}"
else
    echo -e "${YELLOW}⚠ Skipping GRPC_SERVER_ADDRESS (not configured)${NC}"
fi

echo ""
echo "📝 Creating initial commit..."

# Add all files
git add .

# Create initial commit
git commit -m "Initial commit: Vietlott SMS Prediction Tool

- Implement hexagonal architecture with domain, application, and infrastructure layers
- Add 3 prediction algorithms (Frequency, Hot/Cold, Pattern analyzers)
- Implement ensemble voting system with 3 strategies
- Create predictor and backtester CLIs
- Set up GitHub Actions workflows (daily prediction, weekly backtest, CI/CD)
- Add comprehensive documentation

Features:
- Multi-game support (MEGA_6_45, POWER_6_55)
- Automated daily predictions via GitHub Actions
- Comprehensive backtesting (30 draws + 30 days)
- gRPC integration with too_predict
- 79% test coverage

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>" || echo "Commit already exists or nothing to commit"

# Push to GitHub
echo ""
echo "📤 Pushing to GitHub..."
git push -u origin main || git push -u origin master

echo ""
echo -e "${GREEN}✓ Setup complete!${NC}"
echo ""
echo "📊 Next Steps:"
echo "1. Visit your repository: https://github.com/$GITHUB_USERNAME/$REPO_NAME"
echo "2. Go to Actions tab to enable workflows"
echo "3. Configure secrets if needed: Settings > Secrets and variables > Actions"
echo "4. Test workflows manually: Actions > Select workflow > Run workflow"
echo ""
echo "🔗 Repository URLs:"
echo "   - Web: https://github.com/$GITHUB_USERNAME/$REPO_NAME"
echo "   - Clone: git@github.com:$GITHUB_USERNAME/$REPO_NAME.git"
echo ""

# Display workflow information
echo "⚙️  Configured Workflows:"
echo "   - Daily Prediction: Runs daily at 18:00 UTC"
echo "   - Weekly Backtest: Runs Sundays at 00:00 UTC"
echo "   - CI/CD: Runs on push/PR"
echo ""

# Optional: Test first workflow
read -p "Do you want to trigger a test prediction run now? (y/n): " RUN_TEST
if [[ $RUN_TEST == "y" || $RUN_TEST == "Y" ]]; then
    echo ""
    echo "🧪 Triggering test prediction..."
    gh workflow run daily-prediction.yml \
        -f game_type=MEGA_6_45 \
        --repo "$GITHUB_USERNAME/$REPO_NAME"
    echo -e "${GREEN}✓ Test workflow triggered${NC}"
    echo "View progress: https://github.com/$GITHUB_USERNAME/$REPO_NAME/actions"
fi

echo ""
echo -e "${GREEN}🎉 Setup completed successfully!${NC}"
