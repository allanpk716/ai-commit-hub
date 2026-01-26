#!/usr/bin/env python3
"""Prepare multiple PNG icon sizes from source"""

from PIL import Image
import os

def prepare_icons():
    """Generate multiple PNG files for different icon sizes"""
    source_png = "assets/icons/appicon.png"
    output_dir = "winres"

    # Create output directory
    os.makedirs(output_dir, exist_ok=True)

    # Standard icon sizes
    sizes = {
        "icon16.png": 16,
        "icon32.png": 32,
        "icon48.png": 48,
        "icon64.png": 64,
        "icon128.png": 128,
        "icon256.png": 256
    }

    print("Generating icon PNG files...")

    for filename, size in sizes.items():
        output_path = os.path.join(output_dir, filename)
        try:
            img = Image.open(source_png)
            if img.mode != 'RGBA':
                img = img.convert('RGBA')

            resized = img.resize((size, size), Image.Resampling.LANCZOS)
            resized.save(output_path, 'PNG')
            print(f"  Generated: {filename} ({size}x{size})")
        except Exception as e:
            print(f"  ERROR generating {filename}: {e}")

    print(f"\nAll icons generated in {output_dir}/")

if __name__ == "__main__":
    prepare_icons()
