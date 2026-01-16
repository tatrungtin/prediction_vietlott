#!/bin/bash

# Vietlott Crawl and Commit Script
# This script crawls the latest Vietlott data and commits changes to git

set -e

echo "🎲 Starting Vietlott crawl..."
echo ""

# Run the crawler
echo "📡 Fetching data from Vietlott..."
go run scripts/crawl_vietlott.go

echo ""
echo "📊 Checking for changes..."

# Check if there are any changes in the data directory
if [ -n "$(git status --porcelain data/draws/)" ]; then
    echo "✅ New data found!"
    echo ""

    # Show what changed
    echo "📁 Files changed:"
    git status --short data/draws/
    echo ""

    # Add changes
    git add data/draws/

    # Commit with current date
    COMMIT_MESSAGE="chore: crawl Vietlott data - $(date +'%Y-%m-%d %H:%M:%S')"
    git commit -m "$COMMIT_MESSAGE"

    echo ""
    echo "✅ Changes committed successfully!"
    echo "📝 Commit message: $COMMIT_MESSAGE"
    echo ""
    echo "💡 To push to remote, run: git push"

else
    echo "ℹ️  No new data available from Vietlott"
    echo "   All data is up to date"
fi

echo ""
echo "✨ Done!"
