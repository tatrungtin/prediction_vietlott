#!/usr/bin/env python3
"""
Generate additional sample draws to reach 30 total.
This script extends the existing 8 real draws with 22 more sample draws.
"""

import json
import random
from datetime import datetime, timedelta

# Read existing draws to avoid duplicates
existing_draws = set()
with open('data/draws/power_6_55/power_01288.json', 'r') as f:
    data = json.load(f)
    base_date = datetime.fromisoformat(data['draw_date'].replace('Z', '+00:00'))
    base_draw_number = data['draw_number']

# Generate 22 more draws going backwards
start_draw_number = base_draw_number - 1
start_date = base_date - timedelta(days=3)  # Power 6/55 is drawn ~3 times per week

for i in range(22):
    draw_number = start_draw_number - i
    # Calculate date going backwards (Power 6/55 is drawn Tue, Thu, Sat)
    days_back = (i // 3 + 1) * 3  # Approximate 3 days per draw
    draw_date = start_date - timedelta(days=days_back)

    # Generate 6 unique random numbers from 1-55
    numbers = sorted(random.sample(range(1, 56), 6))

    draw = {
        "id": f"power_{draw_number:05d}",
        "game_type": "POWER_6_55",
        "draw_number": draw_number,
        "numbers": numbers,
        "draw_date": draw_date.isoformat(),
        "jackpot": 0,
        "winners": 0
    }

    filename = f"data/draws/power_6_55/power_{draw_number:05d}.json"
    with open(filename, 'w') as f:
        json.dump(draw, f, indent=2)

    print(f"Generated: {filename} - {draw_date.strftime('%d/%m/%Y')} - {numbers}")

print(f"\n✅ Generated 22 additional sample draws")
print(f"📊 Total draws: 8 real + 22 sample = 30 draws")
