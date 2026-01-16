#!/bin/bash

# Initialize sample historical data for POWER_6_55
# Creates individual draw files that the storage system can read

set -e

GAME_TYPE="power_6_55"
BASE_DIR="data/draws/$GAME_TYPE"

echo "📊 Creating sample historical data for POWER_6/55..."

# Create directory
mkdir -p "$BASE_DIR"

# Generate 30 sample draws with realistic patterns
for i in {1..30}; do
  DRAW_NUM=$((1000 - i + 1))
  ID="power_$(printf "%03d" $i)"

  # Generate 6 random unique numbers between 1-55
  NUMBERS=$(shuf -i 1-55 -n 6 | sort -n | paste -sd, -)

  # Calculate date (going back from today)
  DATE=$(date -v-${i}d +%Y-%m-%dT18:00:00Z)

  # Create draw JSON
  cat > "$BASE_DIR/$ID.json" << EOF
{
  "id": "$ID",
  "game_type": "POWER_6_55",
  "draw_number": $DRAW_NUM,
  "numbers": [$NUMBERS],
  "draw_date": "$DATE",
  "jackpot": $((RANDOM % 50000000000 + 1000000000)),
  "winners": $((RANDOM % 3))
}
EOF
done

echo "✅ Created 30 sample draws for POWER_6/55"
echo "📁 Location: $BASE_DIR"
echo ""
echo "📊 Sample files:"
ls -1 "$BASE_DIR" | head -5
echo "..."
echo ""
echo "🎯 Now you can run:"
echo "   ./bin/predictor -g POWER_6_55"
