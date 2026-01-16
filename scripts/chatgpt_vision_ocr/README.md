# ChatGPT Vision OCR for Vietlott Power 6/55

This script uses **ChatGPT's GPT-4o Vision API** to extract winning lottery numbers from Vietlott result PDF images.

## Why ChatGPT Vision?

✅ **More accurate** than traditional OCR - understands context
✅ **Works perfectly** on macOS ARM64 (no compatibility issues)
✅ **Handles complex layouts** - can find numbers in any format
✅ **Cost effective** - ~$0.01 per image (~$0.50 for 50 images)

## Prerequisites

### 1. Get OpenAI API Key

1. Go to: https://platform.openai.com/api-keys
2. Create a new API key
3. Copy your key (starts with `sk-proj-...`)

### 2. Install Dependencies

```bash
# Install OpenAI Python library
pip3 install openai

# Or using pip3 directly
python3 -m pip install openai
```

## Usage

### Quick Test (1 Image)

```bash
# Set your API key
export OPENAI_API_KEY='sk-proj-your-api-key-here'

# Run script in test mode
python3 scripts/chatgpt_vision_ocr/chatgpt_ocr.py

# Enter '0' when prompted to test with 1 image
```

### Process All 50 Images

```bash
# Set your API key
export OPENAI_API_KEY='sk-proj-your-api-key-here'

# Run script
python3 scripts/chatgpt_vision_ocr/chatgpt_ocr.py

# Press Enter to process all images
```

### Process Specific Number of Images

```bash
python3 scripts/chatgpt_vision_ocr/chatgpt_ocr.py

# Enter number: 10 (processes first 10 images)
```

## How It Works

1. **Reads converted images** from `/tmp/vietlott_images/` (already converted from PDFs)
2. **Encodes image** to base64
3. **Sends to ChatGPT Vision API** (GPT-4o model)
4. **Extracts 6 winning numbers** from the response
5. **Saves as JSON** in `data/draws/power_6_55/`

## Expected Output

```
============================================================
Vietlott Power 6/55 - ChatGPT Vision OCR Crawler
============================================================

✓ OpenAI client initialized
✓ Found 8 existing draws

How many images would you like to process?
  - Enter a number (1-50) to process that many images
  - Press Enter to process all 50 images
  - Enter '0' to test with just 1 image

Number of images [0=test, 1=1 image, Enter=all]: 1

Testing with 1 image...

=== Processing Images with ChatGPT Vision ===

  [01237] Analyzing with ChatGPT Vision...
  [01237] ChatGPT Response: The winning numbers are: 05, 12, 23, 34, 42, 55
  ✓ Saved: [5, 12, 23, 34, 42, 55]

=== Summary ===
Processed: 1 images
Saved: 1 new draws
Failed: 0 draws

API Cost: ~$0.01 USD
```

## API Cost

- **GPT-4o Vision**: ~$0.01 per image
- **50 images**: ~$0.50 USD
- **Test with 1 image**: ~$0.01 USD

## Troubleshooting

### Error: "OpenAI API key not found"

```bash
# Set your API key
export OPENAI_API_KEY='sk-proj-your-api-key-here'

# Or add to your shell profile (~/.zshrc or ~/.bash_profile)
echo 'export OPENAI_API_KEY="sk-proj-your-api-key-here"' >> ~/.zshrc
source ~/.zshrc
```

### Error: "No module named 'openai'"

```bash
pip3 install openai
```

### Images Not Found

```bash
# Check if images exist
ls /tmp/vietlott_images/*.png

# If empty, run the PDF crawler first to generate images
```

## After Extraction

Once you have extracted the numbers:

```bash
# Check your new draws
ls -1 data/draws/power_6_55/*.json | wc -l

# Run predictor with expanded dataset
./bin/predictor -g POWER_6_55

# Run backtester with more data
./bin/backtester -g POWER_6_55 -m draws -s 30
```

## Advantages Over Traditional OCR

| Feature | ChatGPT Vision | Traditional OCR |
|---------|----------------|-----------------|
| Accuracy | ⭐⭐⭐⭐⭐ Excellent | ⭐⭐⭐ Variable |
| Layout handling | Understands context | Needs templates |
| macOS ARM64 | ✅ Works perfectly | ❌ Compatibility issues |
| Complex images | ✅ Handles well | ❌ Struggles |
| Setup | Easy (API key) | Complex (install dependencies) |

## Next Steps

1. **Test with 1 image** to verify it works
2. **Check API balance** at: https://platform.openai.com/usage
3. **Process all 50 images** (~$0.50 total cost)
4. **Verify extracted numbers** before using in predictions

## Example Extracted Data

```json
{
  "id": "power_01237",
  "game_type": "POWER_6_55",
  "draw_number": 1237,
  "numbers": [5, 12, 23, 34, 42, 55],
  "draw_date": "2025-11-04T18:00:00Z"
}
```
