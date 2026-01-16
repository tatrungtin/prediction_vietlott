#!/bin/bash

# Create sample historical data for testing

set -e

echo "📊 Creating sample historical data for POWER_6_55..."

# Create data directory
mkdir -p data/draws/power_6_55

# Generate 30 sample draws (most recent first)
cat > data/draws/power_6_55/sample_draws.json << 'EOF'
[
  {"id":"power_001","game_type":"POWER_6_55","draw_number":1000,"numbers":[5,12,23,31,42,50],"draw_date":"2026-01-15T18:00:00Z","jackpot":50000000000,"winners":0},
  {"id":"power_002","game_type":"POWER_6_55","draw_number":999,"numbers":[8,15,22,33,41,48],"draw_date":"2026-01-14T18:00:00Z","jackpot":45000000000,"winners":0},
  {"id":"power_003","game_type":"POWER_6_55","draw_number":998,"numbers":[3,11,19,27,38,52],"draw_date":"2026-01-13T18:00:00Z","jackpot":40000000000,"winners":0},
  {"id":"power_004","game_type":"POWER_6_55","draw_number":997,"numbers":[7,14,25,32,44,49],"draw_date":"2026-01-12T18:00:00Z","jackpot":35000000000,"winners":1},
  {"id":"power_005","game_type":"POWER_6_55","draw_number":996,"numbers":[2,18,26,34,43,51],"draw_date":"2026-01-11T18:00:00Z","jackpot":30000000000,"winners":0},
  {"id":"power_006","game_type":"POWER_6_55","draw_number":995,"numbers":[9,16,21,35,40,47],"draw_date":"2026-01-10T18:00:00Z","jackpot":25000000000,"winners":0},
  {"id":"power_007","game_type":"POWER_6_55","draw_number":994,"numbers":[4,13,28,36,45,53],"draw_date":"2026-01-09T18:00:00Z","jackpot":20000000000,"winners":0},
  {"id":"power_008","game_type":"POWER_6_55","draw_number":993,"numbers":[6,17,24,31,39,50],"draw_date":"2026-01-08T18:00:00Z","jackpot":15000000000,"winners":0},
  {"id":"power_009","game_type":"POWER_6_55","draw_number":992,"numbers":[10,19,29,37,46,54],"draw_date":"2026-01-07T18:00:00Z","jackpot":10000000000,"winners":2},
  {"id":"power_010","game_type":"POWER_6_55","draw_number":991,"numbers":[1,14,23,32,41,48],"draw_date":"2026-01-06T18:00:00Z","jackpot":8000000000,"winners":0},
  {"id":"power_011","game_type":"POWER_6_55","draw_number":990,"numbers":[5,18,27,35,44,52],"draw_date":"2026-01-05T18:00:00Z","jackpot":7000000000,"winners":0},
  {"id":"power_012","game_type":"POWER_6_55","draw_number":989,"numbers":[8,16,25,33,42,49],"draw_date":"2026-01-04T18:00:00Z","jackpot":6000000000,"winners":0},
  {"id":"power_013","game_type":"POWER_6_55","draw_number":988,"numbers":[3,12,21,30,38,47],"draw_date":"2026-01-03T18:00:00Z","jackpot":5000000000,"winners":1},
  {"id":"power_014","game_type":"POWER_6_55","draw_number":987,"numbers":[7,15,26,34,43,51],"draw_date":"2026-01-02T18:00:00Z","jackpot":4000000000,"winners":0},
  {"id":"power_015","game_type":"POWER_6_55","draw_number":986,"numbers":[2,19,28,36,45,53],"draw_date":"2026-01-01T18:00:00Z","jackpot":3000000000,"winners":0},
  {"id":"power_016","game_type":"POWER_6_55","draw_number":985,"numbers":[9,17,24,31,40,48],"draw_date":"2025-12-31T18:00:00Z","jackpot":2000000000,"winners":0},
  {"id":"power_017","game_type":"POWER_6_55","draw_number":984,"numbers":[4,13,22,32,39,50],"draw_date":"2025-12-30T18:00:00Z","jackpot":1500000000,"winners":1},
  {"id":"power_018","game_type":"POWER_6_55","draw_number":983,"numbers":[6,18,29,37,46,54],"draw_date":"2025-12-29T18:00:00Z","jackpot":1000000000,"winners":0},
  {"id":"power_019","game_type":"POWER_6_55","draw_number":982,"numbers":[11,20,25,35,44,52],"draw_date":"2025-12-28T18:00:00Z","jackpot":900000000,"winners":0},
  {"id":"power_020","game_type":"POWER_6_55","draw_number":981,"numbers":[1,14,27,33,42,49],"draw_date":"2025-12-27T18:00:00Z","jackpot":800000000,"winners":0},
  {"id":"power_021","game_type":"POWER_6_55","draw_number":980,"numbers":[5,16,23,30,41,47],"draw_date":"2025-12-26T18:00:00Z","jackpot":700000000,"winners":0},
  {"id":"power_022","game_type":"POWER_6_55","draw_number":979,"numbers":[8,19,26,34,43,51],"draw_date":"2025-12-25T18:00:00Z","jackpot":600000000,"winners":2},
  {"id":"power_023","game_type":"POWER_6_55","draw_number":978,"numbers":[3,12,21,38,45,53],"draw_date":"2025-12-24T18:00:00Z","jackpot":500000000,"winners":0},
  {"id":"power_024","game_type":"POWER_6_55","draw_number":977,"numbers":[7,15,28,36,40,48],"draw_date":"2025-12-23T18:00:00Z","jackpot":400000000,"winners":0},
  {"id":"power_025","game_type":"POWER_6_55","draw_number":976,"numbers":[2,17,24,31,42,50],"draw_date":"2025-12-22T18:00:00Z","jackpot":300000000,"winners":0},
  {"id":"power_026","game_type":"POWER_6_55","draw_number":975,"numbers":[9,20,29,37,46,54],"draw_date":"2025-12-21T18:00:00Z","jackpot":200000000,"winners":1},
  {"id":"power_027","game_type":"POWER_6_55","draw_number":974,"numbers":[4,13,22,32,39,52],"draw_date":"2025-12-20T18:00:00Z","jackpot":150000000,"winners":0},
  {"id":"power_028","game_type":"POWER_6_55","draw_number":973,"numbers":[6,18,25,33,44,49],"draw_date":"2025-12-19T18:00:00Z","jackpot":100000000,"winners":0},
  {"id":"power_029","game_type":"POWER_6_55","draw_number":972,"numbers":[10,19,27,35,41,47],"draw_date":"2025-12-18T18:00:00Z","jackpot":90000000,"winners":0},
  {"id":"power_030","game_type":"POWER_6_55","draw_number":971,"numbers":[1,14,23,30,43,51],"draw_date":"2025-12-17T18:00:00Z","jackpot":80000000,"winners":0}
]
EOF

echo "✅ Sample data created for POWER_6_55"
echo ""
echo "📁 Sample draws saved to: data/draws/power_6_55/sample_draws.json"
echo ""
echo "💡 Now you can:"
echo "   1. Import this data manually to storage"
echo "   2. Run predictor to test with sample data"
