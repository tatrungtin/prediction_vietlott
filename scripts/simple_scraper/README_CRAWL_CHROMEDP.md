# Headless Browser Crawler for Vietlott Power 6/55

This crawler uses **chromedp** (Chrome DevTools Protocol) to navigate JavaScript-heavy pages on the Vietlott website and extract historical draw data.

## What it does

1. **Navigates through 5 pages** of the announcement section (thong-bao-ket-qua-655)
2. **Extracts draw info** (draw number, date) from each page's table
3. **Fetches winning numbers** from individual draw detail pages (requires JavaScript execution)
4. **Saves to JSON files** in `data/draws/power_6_55/`
5. **Skips duplicates** - won't overwrite existing draws

## Prerequisites

### 1. Install Chrome/Chromium

The crawler requires Google Chrome or Chromium browser installed:

```bash
# macOS
brew install --criterion google-chrome

# Linux (Ubuntu/Debian)
sudo apt-get install chromium-browser

# Or download Chrome from: https://www.google.com/chrome/
```

### 2. Install Go dependencies

```bash
cd /Users/tinta/Documents/go_project/tool_predict
go get github.com/chromedp/chromedp
```

## Running the Crawler

```bash
# From project root directory
cd /Users/tinta/Documents/go_project/tool_predict

# Run the crawler
go run scripts/crawl_chromedp.go
```

## Expected Output

The crawler will:
- Process 5 pages of announcements
- Extract ~50 draws total (10 per page)
- Save new draws as JSON files: `power_00678.json`, `power_00679.json`, etc.
- Show progress: "Crawling announcement page 1/5...", "Fetching numbers for draw 687..."

## Output Format

Each draw is saved as a JSON file:

```json
{
  "id": "power_00687",
  "game_type": "POWER_6_55",
  "draw_number": 687,
  "numbers": [5, 12, 23, 31, 42, 55],
  "draw_date": "2022-02-19T18:00:00Z",
  "jackpot": 0,
  "winners": 0
}
```

## Limitations

- **Speed**: Headless browser is slower than HTTP scraping (~2-3 seconds per page)
- **Data Age**: This crawler fetches OLD data from 2022 (February), not current draws
- **JavaScript Required**: Vietlott website uses JavaScript to load winning numbers

## Alternative: Use Current Data Only

If you only want to work with current/recent data (your existing 8 draws from Dec 2025 - Jan 2026), continue using the daily crawler at:

```
.github/workflows/daily-prediction.yml
```

This will gradually accumulate recent data over time.

## Troubleshooting

### Chrome not found

If you get an error like "executable file not found", install Chrome:

```bash
# macOS
brew install --criterion google-chrome

# Linux
sudo apt-get install chromium-browser
```

### Timeout errors

Increase the timeout in `crawl_chromedp.go`:

```go
ctx, cancel = context.WithTimeout(ctx, 10*time.Minute) // Increase from 5 to 10 minutes
```

### No numbers extracted

The headless browser may need more time to load JavaScript. Increase the sleep time:

```go
chromedp.Sleep(3*time.Second), // Increase from 2 to 3 seconds
```

## Next Steps After Crawling

After successfully crawling the historical data:

1. **Verify the data**:
   ```bash
   ls -la data/draws/power_6_55/
   cat data/draws/power_6_55/power_00687.json
   ```

2. **Run the predictor** with the expanded dataset:
   ```bash
   ./bin/predictor -g POWER_6_55
   ```

3. **Run backtesting** with more historical data:
   ```bash
   ./bin/backtester -g POWER_6_55 -m draws -s 30
   ```

4. **Check algorithm performance**:
   - Frequency Analyzer: Works better with 50+ draws
   - Hot/Cold Analyzer: Still needs 50+ draws
   - Pattern Analyzer: Still needs 100+ draws
