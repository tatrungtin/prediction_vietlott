#!/usr/bin/env python3
"""
Vietlott Power 6/55 OCR Crawler using ChatGPT Vision API
Extracts winning numbers from lottery result images using GPT-4o Vision
"""

import os
import base64
import re
import json
import time
import requests
from pathlib import Path
from typing import List, Optional
from datetime import datetime
from openai import OpenAI

# Configuration
PDF_DIR = "/tmp/vietlott_pdfs"
IMAGE_DIR = "/tmp/vietlott_images"
OUTPUT_DIR = "data/draws/power_6_55"
BASE_URL = "https://vietlott.vn/vi/trung-thuong/ket-qua-trung-thuong/thong-bao-ket-qua-655?pageindex={}&nocatche=1"
TOTAL_PAGES = 5

class VietlottChatGPTOCR:
    def __init__(self, api_key: str = None):
        """
        Initialize with OpenAI API key
        Get your API key from: https://platform.openai.com/api-keys
        """
        if api_key:
            self.client = OpenAI(api_key=api_key)
        else:
            # Try to get from environment variable
            api_key = os.environ.get("OPENAI_API_KEY")
            if not api_key:
                raise Exception(
                    "OpenAI API key not found!\n"
                    "Get your API key from: https://platform.openai.com/api-keys\n"
                    "Then set it as: export OPENAI_API_KEY='your-api-key-here'"
                )
            self.client = OpenAI(api_key=api_key)

        self.existing_draws = self.get_existing_draws()
        print(f"✓ OpenAI client initialized")
        print(f"✓ Found {len(self.existing_draws)} existing draws")

    def get_existing_draws(self) -> set:
        """Get set of existing draw IDs"""
        existing = set()
        output_path = Path(OUTPUT_DIR)
        if output_path.exists():
            for file in output_path.glob("power_*.json"):
                existing.add(file.stem)
        return existing

    def encode_image(self, image_path: str) -> str:
        """Encode image to base64"""
        with open(image_path, "rb") as image_file:
            return base64.b64encode(image_file.read()).decode('utf-8')

    def extract_numbers_with_chatgpt(self, image_path: str, draw_number: int) -> tuple[List[int], Optional[datetime]]:
        """Extract winning numbers and date using ChatGPT Vision API"""

        print(f"  [{draw_number:05d}] Analyzing with ChatGPT Vision...")

        # Read and encode image
        base64_image = self.encode_image(image_path)

        # Call GPT-4o Vision API
        response = self.client.chat.completions.create(
            model="gpt-4o",  # GPT-4o has excellent vision capabilities
            messages=[
                {
                    "role": "user",
                    "content": [
                        {
                            "type": "text",
                            "text": (
                                "This is a Vietlott Power 6/55 lottery result image. "
                                "Please extract:\n"
                                "1. The draw date (ngày mở thưởng) - shown in DD/MM/YYYY format\n"
                                "2. The 6 winning numbers\n\n"
                                "The numbers are typically displayed as 6 two-digit numbers ranging from 01 to 55.\n"
                                "Look for numbers that are often:\n"
                                "- Displayed in a grid or circle pattern\n"
                                "- Separated by dashes, spaces, or arranged vertically\n"
                                "- Sometimes labeled as 'Bộ số' (winning numbers)\n\n"
                                "Return in this exact format:\n"
                                "Date: DD/MM/YYYY\n"
                                "Numbers: 01, 15, 23, 34, 42, 55"
                            )
                        },
                        {
                            "type": "image_url",
                            "image_url": {
                                "url": f"data:image/png;base64,{base64_image}"
                            }
                        }
                    ]
                }
            ],
            max_tokens=300,
            temperature=0
        )

        # Extract response text
        ocr_text = response.choices[0].message.content.strip()
        print(f"  [{draw_number:05d}] ChatGPT Response: {ocr_text}")

        # Parse date and numbers from response
        draw_date = self.parse_date_from_response(ocr_text)
        numbers = self.parse_numbers_from_response(ocr_text)

        return numbers, draw_date

    def parse_date_from_response(self, text: str) -> Optional[datetime]:
        """Parse draw date from ChatGPT response"""
        # Look for date pattern DD/MM/YYYY
        date_match = re.search(r'Date:\s*(\d{2})/(\d{2})/(\d{4})', text)
        if date_match:
            day, month, year = map(int, date_match.groups())
            return datetime(year, month, day, 18, 0, 0)
        return None

    def parse_numbers_from_response(self, text: str) -> List[int]:
        """Parse 6 numbers from ChatGPT response"""

        # Remove common separators and split
        text = text.replace('-', ' ').replace(',', ' ').replace('  ', ' ')
        parts = text.split()

        numbers = []
        for part in parts:
            # Try to extract number from each part
            match = re.search(r'\d+', part)
            if match:
                num = int(match.group())
                if 1 <= num <= 55 and num not in numbers:
                    numbers.append(num)
                    if len(numbers) == 6:
                        break

        # Sort the numbers
        numbers.sort()

        return numbers

    def convert_pdf_to_image(self, pdf_path: str, draw_number: int) -> str:
        """Convert PDF to image using pdftoppm"""
        image_base = os.path.join(IMAGE_DIR, f"draw_{draw_number:05d}")
        image_path = f"{image_base}.png"

        if os.path.exists(image_path):
            return image_path

        # Use pdftoppm to convert
        import subprocess
        cmd = ["pdftoppm", "-png", "-singlefile", pdf_path, image_base]
        result = subprocess.run(cmd, capture_output=True, text=True)

        if result.returncode != 0:
            raise Exception(f"pdftoppm failed: {result.stderr}")

        if not os.path.exists(image_path):
            raise Exception(f"Image not created: {image_path}")

        return image_path

    def save_draw(self, draw_number: int, numbers: List[int], draw_date: datetime):
        """Save draw to JSON file"""
        draw = {
            "id": f"power_{draw_number:05d}",
            "game_type": "POWER_6_55",
            "draw_number": draw_number,
            "numbers": numbers,
            "draw_date": draw_date.isoformat() + "Z",
        }

        filename = f"power_{draw_number:05d}.json"
        filepath = os.path.join(OUTPUT_DIR, filename)

        with open(filepath, 'w') as f:
            json.dump(draw, f, indent=2)

        print(f"  ✓ Saved: {numbers}")

    def process_existing_images(self, limit: int = None):
        """Process already converted images with ChatGPT Vision"""
        print("\n=== Processing Images with ChatGPT Vision ===\n")

        os.makedirs(OUTPUT_DIR, exist_ok=True)

        # Get all PNG images
        image_files = sorted(Path(IMAGE_DIR).glob("draw_*.png"))

        if limit:
            image_files = image_files[:limit]

        total_draws = 0
        saved_draws = 0
        failed_draws = 0

        for image_file in image_files:
            draw_number = int(image_file.stem.split("_")[1])

            draw_id = f"power_{draw_number:05d}"
            if draw_id in self.existing_draws:
                print(f"  [{draw_number:05d}] Updating existing draw (fixing date)...")

            try:
                # Extract numbers and date using ChatGPT Vision
                numbers, draw_date = self.extract_numbers_with_chatgpt(str(image_file), draw_number)

                if len(numbers) == 6 and draw_date is not None:
                    self.save_draw(draw_number, numbers, draw_date)
                    saved_draws += 1
                else:
                    if draw_date is None:
                        print(f"  [{draw_number:05d}] ⚠ Could not extract date")
                    if len(numbers) != 6:
                        print(f"  [{draw_number:05d}] ⚠ Could not extract 6 numbers (got {len(numbers)})")
                    failed_draws += 1

                total_draws += 1

                # Rate limiting - avoid hitting API limits
                time.sleep(1)

            except Exception as e:
                print(f"  [{draw_number:05d}] ✗ Error: {e}")
                failed_draws += 1
                continue

        print(f"\n=== Summary ===")
        print(f"Processed: {total_draws} images")
        print(f"Saved: {saved_draws} new draws")
        print(f"Failed: {failed_draws} draws")
        print(f"\nAPI Cost: ~${saved_draws * 0.01:.2f} USD (GPT-4o Vision: ~$0.01 per image)")

def main():
    """Main entry point"""
    print("=" * 60)
    print("Vietlott Power 6/55 - ChatGPT Vision OCR Crawler")
    print("=" * 60)
    print()

    try:
        ocr = VietlottChatGPTOCR()

        # Ask user how many images to process
        print("\nHow many images would you like to process?")
        print("  - Enter a number (1-50) to process that many images")
        print("  - Press Enter to process all 50 images")
        print("  - Enter '0' to test with just 1 image")

        user_input = input("\nNumber of images [0=test, 1=1 image, Enter=all]: ").strip()

        if user_input == "" or not user_input:
            limit = None  # Process all
            print(f"\nProcessing all 50 images...")
        else:
            limit = int(user_input)
            if limit == 0:
                limit = 1
                print(f"\nTesting with 1 image...")
            else:
                print(f"\nProcessing {limit} images...")

        ocr.process_existing_images(limit=limit)

    except Exception as e:
        print(f"\n✗ Error: {e}")
        print("\nTroubleshooting:")
        print("1. Make sure you have set your OpenAI API key:")
        print("   export OPENAI_API_KEY='sk-proj-...'")
        print("2. Get your API key from: https://platform.openai.com/api-keys")
        print("3. Install required packages:")
        print("   pip3 install openai")

if __name__ == "__main__":
    main()
