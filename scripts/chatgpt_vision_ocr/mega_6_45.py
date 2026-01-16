#!/usr/bin/env python3
"""
Vietlott Mega 6/45 OCR Crawler using ChatGPT Vision API
Downloads PDFs, converts to images, and extracts winning numbers using GPT-4o Vision
"""

import os
import base64
import re
import json
import time
import requests
import subprocess
from pathlib import Path
from typing import List, Optional, Tuple
from datetime import datetime
from openai import OpenAI

# Configuration
PDF_DIR = "/tmp/vietlott_pdfs_mega"
IMAGE_DIR = "/tmp/vietlott_images_mega"
OUTPUT_DIR = "data/draws/mega_6_45"
BASE_URL = "https://vietlott.vn/vi/trung-thuong/ket-qua-trung-thuong/thong-bao-ket-qua-645?pageindex={}&nocatche=1"
TOTAL_PAGES = 5  # Adjust to get more draws

class VietlottMegaOCR:
    def __init__(self, api_key: str = None):
        """Initialize with OpenAI API key"""
        if api_key:
            self.client = OpenAI(api_key=api_key)
        else:
            api_key = os.environ.get("OPENAI_API_KEY")
            if not api_key:
                raise Exception(
                    "OpenAI API key not found!\n"
                    "Get your API key from: https://platform.openai.com/api-keys\n"
                    "Then set it as: export OPENAI_API_KEY='your-api-key-here'"
                )
            self.client = OpenAI(api_key=api_key)

        # Create directories
        os.makedirs(PDF_DIR, exist_ok=True)
        os.makedirs(IMAGE_DIR, exist_ok=True)
        os.makedirs(OUTPUT_DIR, exist_ok=True)

        self.existing_draws = self.get_existing_draws()
        print(f"✓ OpenAI client initialized")
        print(f"✓ Found {len(self.existing_draws)} existing draws")

    def get_existing_draws(self) -> set:
        """Get set of existing draw IDs"""
        existing = set()
        output_path = Path(OUTPUT_DIR)
        if output_path.exists():
            for file in output_path.glob("mega_*.json"):
                existing.add(file.stem)
        return existing

    def encode_image(self, image_path: str) -> str:
        """Encode image to base64"""
        with open(image_path, "rb") as image_file:
            return base64.b64encode(image_file.read()).decode('utf-8')

    def fetch_pdf_links(self) -> List[Tuple[int, str, datetime]]:
        """Fetch all PDF links from Vietlott announcement pages"""
        print(f"\n=== Fetching PDF Links ===\n")
        all_pdfs = []

        for page in range(1, TOTAL_PAGES + 1):
            print(f"Fetching page {page}/{TOTAL_PAGES}...")

            url = BASE_URL.format(page)
            response = requests.get(url, headers={
                'User-Agent': 'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36'
            }, timeout=30)

            if response.status_code != 200:
                print(f"  ✗ Failed to fetch page {page}: status {response.status_code}")
                continue

            # Extract PDF links from HTML
            pdf_pattern = r'https://media\.vietlott\.vn/[^\s]*\[645\][^\s]*\.pdf'
            pdf_matches = re.findall(pdf_pattern, response.text)

            # Extract draw numbers and dates from the page
            draw_pattern = r'\[645\]---(\d+)---'
            date_pattern = r'Ngày\s+(\d{2})/(\d{2})/(\d{4})'

            for pdf_url in pdf_matches:
                draw_match = re.search(draw_pattern, pdf_url)
                if not draw_match:
                    continue

                draw_number = int(draw_match.group(1))

                # Try to find date near this PDF
                # Find the context around this PDF URL
                pdf_snippet = response.text[response.text.find(pdf_url)-200:response.text.find(pdf_url)+200]
                date_match = re.search(date_pattern, pdf_snippet)

                if date_match:
                    day, month, year = map(int, date_match.groups())
                    draw_date = datetime(year, month, day, 18, 0, 0)
                else:
                    draw_date = None

                all_pdfs.append((draw_number, pdf_url, draw_date))

            print(f"  Found {len(pdf_matches)} PDFs on page {page}")
            time.sleep(1)

        print(f"\nTotal PDFs found: {len(all_pdfs)}")
        return all_pdfs

    def download_pdf(self, draw_number: int, pdf_url: str) -> Optional[str]:
        """Download PDF from URL"""
        pdf_path = os.path.join(PDF_DIR, f"draw_{draw_number:05d}.pdf")

        if os.path.exists(pdf_path):
            return pdf_path

        try:
            print(f"  [{draw_number:05d}] Downloading PDF...")
            response = requests.get(pdf_url, headers={
                'User-Agent': 'Mozilla/5.0 (compatible; tool_predict/1.0)'
            }, timeout=30)

            if response.status_code == 200:
                with open(pdf_path, 'wb') as f:
                    f.write(response.content)
                return pdf_path
            else:
                print(f"  [{draw_number:05d}] ✗ Failed to download PDF: status {response.status_code}")
                return None
        except Exception as e:
            print(f"  [{draw_number:05d}] ✗ Error downloading PDF: {e}")
            return None

    def convert_pdf_to_image(self, pdf_path: str, draw_number: int) -> Optional[str]:
        """Convert PDF to PNG image using pdftoppm"""
        image_base = os.path.join(IMAGE_DIR, f"draw_{draw_number:05d}")
        image_path = f"{image_base}.png"

        if os.path.exists(image_path):
            return image_path

        try:
            # Use pdftoppm to convert
            cmd = ["pdftoppm", "-png", "-singlefile", pdf_path, image_base]
            result = subprocess.run(cmd, capture_output=True, text=True)

            if result.returncode != 0:
                print(f"  [{draw_number:05d}] ✗ pdftoppm failed: {result.stderr}")
                return None

            if not os.path.exists(image_path):
                print(f"  [{draw_number:05d}] ✗ Image not created")
                return None

            return image_path
        except Exception as e:
            print(f"  [{draw_number:05d}] ✗ Error converting PDF: {e}")
            return None

    def extract_numbers_with_chatgpt(self, image_path: str, draw_number: int) -> Tuple[List[int], Optional[datetime]]:
        """Extract winning numbers and date using ChatGPT Vision API"""

        print(f"  [{draw_number:05d}] Analyzing with ChatGPT Vision...")

        base64_image = self.encode_image(image_path)

        response = self.client.chat.completions.create(
            model="gpt-4o",
            messages=[
                {
                    "role": "user",
                    "content": [
                        {
                            "type": "text",
                            "text": (
                                "This is a Vietlott Mega 6/45 lottery result image. "
                                "Please extract:\n"
                                "1. The draw date (ngày mở thưởng) - shown in DD/MM/YYYY format\n"
                                "2. The 6 winning numbers\n\n"
                                "The numbers range from 01 to 45 (not 55!). "
                                "Look for numbers that are often:\n"
                                "- Displayed in a grid or circle pattern\n"
                                "- Separated by dashes, spaces, or arranged vertically\n"
                                "- Sometimes labeled as 'Bộ số' (winning numbers)\n\n"
                                "IMPORTANT: Only extract numbers between 01-45. If you see numbers like 46-55, "
                                "those are from Power 6/55, not Mega 6/45. Look for the 6 main numbers.\n\n"
                                "Return in this exact format:\n"
                                "Date: DD/MM/YYYY\n"
                                "Numbers: 01, 15, 23, 34, 42, 45"
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

        ocr_text = response.choices[0].message.content.strip()
        print(f"  [{draw_number:05d}] ChatGPT Response: {ocr_text[:100]}...")

        draw_date = self.parse_date_from_response(ocr_text)
        numbers = self.parse_numbers_from_response(ocr_text)

        return numbers, draw_date

    def parse_date_from_response(self, text: str) -> Optional[datetime]:
        """Parse draw date from ChatGPT response"""
        date_match = re.search(r'Date:\s*(\d{2})/(\d{2})/(\d{4})', text)
        if date_match:
            day, month, year = map(int, date_match.groups())
            return datetime(year, month, day, 18, 0, 0)
        return None

    def parse_numbers_from_response(self, text: str) -> List[int]:
        """Parse 6 numbers from ChatGPT response"""
        text = text.replace('-', ' ').replace(',', ' ').replace('  ', ' ')
        parts = text.split()

        numbers = []
        for part in parts:
            match = re.search(r'\d+', part)
            if match:
                num = int(match.group())
                if 1 <= num <= 45 and num not in numbers:  # Mega 6/45 range
                    numbers.append(num)
                    if len(numbers) == 6:
                        break

        numbers.sort()
        return numbers

    def save_draw(self, draw_number: int, numbers: List[int], draw_date: datetime):
        """Save draw to JSON file"""
        draw = {
            "id": f"mega_{draw_number:05d}",
            "game_type": "MEGA_6_45",
            "draw_number": draw_number,
            "numbers": numbers,
            "draw_date": draw_date.isoformat() + "Z",
        }

        filename = f"mega_{draw_number:05d}.json"
        filepath = os.path.join(OUTPUT_DIR, filename)

        with open(filepath, 'w') as f:
            json.dump(draw, f, indent=2)

        print(f"  ✓ Saved: {numbers}")

    def process_all_draws(self, limit: int = None):
        """Main workflow to download, convert, and OCR all draws"""
        print("\n=== Starting Mega 6/45 OCR Pipeline ===\n")

        # Fetch all PDF links
        pdfs = self.fetch_pdf_links()

        if limit:
            pdfs = pdfs[:limit]

        total_draws = 0
        saved_draws = 0
        failed_draws = 0

        for draw_number, pdf_url, known_date in pdfs:
            draw_id = f"mega_{draw_number:05d}"

            if draw_id in self.existing_draws:
                print(f"  [{draw_number:05d}] Skipping (already exists)")
                continue

            try:
                # Download PDF
                pdf_path = self.download_pdf(draw_number, pdf_url)
                if not pdf_path:
                    failed_draws += 1
                    continue

                # Convert to image
                image_path = self.convert_pdf_to_image(pdf_path, draw_number)
                if not image_path:
                    failed_draws += 1
                    continue

                # Extract numbers with ChatGPT Vision
                numbers, draw_date = self.extract_numbers_with_chatgpt(image_path, draw_number)

                # Use known date if ChatGPT couldn't extract it
                if draw_date is None and known_date:
                    draw_date = known_date

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

                # Rate limiting
                time.sleep(1.5)

            except Exception as e:
                print(f"  [{draw_number:05d}] ✗ Error: {e}")
                failed_draws += 1
                continue

        print(f"\n=== Summary ===")
        print(f"Processed: {total_draws} draws")
        print(f"Saved: {saved_draws} new draws")
        print(f"Failed: {failed_draws} draws")
        print(f"\nAPI Cost: ~${saved_draws * 0.01:.2f} USD (GPT-4o Vision: ~$0.01 per image)")

def main():
    print("=" * 60)
    print("Vietlott Mega 6/45 - ChatGPT Vision OCR Crawler")
    print("=" * 60)
    print()

    try:
        ocr = VietlottMegaOCR()

        print("\nHow many draws would you like to process?")
        print("  - Enter a number to process that many draws")
        print("  - Press Enter to process all draws")
        print("  - Enter '0' to test with just 1 draw")

        user_input = input("\nNumber of draws [0=test, 1=1 draw, Enter=all]: ").strip()

        if user_input == "" or not user_input:
            limit = None
            print(f"\nProcessing all draws...")
        else:
            limit = int(user_input)
            if limit == 0:
                limit = 1
                print(f"\nTesting with 1 draw...")
            else:
                print(f"\nProcessing {limit} draws...")

        ocr.process_all_draws(limit=limit)

    except Exception as e:
        print(f"\n✗ Error: {e}")
        print("\nTroubleshooting:")
        print("1. Make sure you have set your OpenAI API key:")
        print("   export OPENAI_API_KEY='sk-proj-...'")
        print("2. Get your API key from: https://platform.openai.com/api-keys")
        print("3. Install required packages:")
        print("   pip3 install openai requests")

if __name__ == "__main__":
    main()