#!/usr/bin/env python3
"""Check and fix Windows ICO file with multiple sizes"""

from PIL import Image
import os

def check_icon_sizes(icon_path):
    """Check what sizes are in the ICO file"""
    try:
        img = Image.open(icon_path)
        sizes = []
        for size in img.info.get('sizes', []):
            sizes.append(size)
        return sizes
    except Exception as e:
        print(f"Error: {e}")
        return []

def generate_multi_size_ico(source_png, output_ico):
    """Generate ICO file with multiple sizes"""
    try:
        # Standard Windows icon sizes
        sizes = [16, 32, 48, 64, 128, 256]

        images = []
        for size in sizes:
            try:
                img = Image.open(source_png)
                # Convert to RGBA if needed
                if img.mode != 'RGBA':
                    img = img.convert('RGBA')

                # Resize with high quality
                resized = img.resize((size, size), Image.Resampling.LANCZOS)
                images.append(resized)
                print(f"  Generated {size}x{size}")
            except Exception as e:
                print(f"  Warning: Could not generate {size}x{size}: {e}")

        # Save as ICO with all sizes
        images[0].save(
            output_ico,
            format='ICO',
            sizes=[(img.width, img.height) for img in images]
        )
        print(f"\n[OK] Generated multi-size ICO: {output_ico}")
        print(f"  Sizes included: {sizes}")
        return True
    except Exception as e:
        print(f"[ERROR] Error generating ICO: {e}")
        return False

if __name__ == "__main__":
    icon_path = "build/windows/icon.ico"
    source_png = "assets/icons/appicon.png"

    print("=" * 50)
    print("Windows ICO File Checker")
    print("=" * 50)

    # Check current icon
    print(f"\n1. Checking current icon: {icon_path}")
    if os.path.exists(icon_path):
        sizes = check_icon_sizes(icon_path)
        if sizes:
            print(f"   Current sizes: {sizes}")
        else:
            print("   Could not read sizes (might be corrupted)")
    else:
        print("   Icon file not found!")

    # Generate new multi-size icon
    print(f"\n2. Generating new multi-size icon from: {source_png}")
    if os.path.exists(source_png):
        generate_multi_size_ico(source_png, icon_path)
    else:
        print(f"[ERROR] Source PNG not found: {source_png}")

    # Verify new icon
    print(f"\n3. Verifying new icon: {icon_path}")
    if os.path.exists(icon_path):
        sizes = check_icon_sizes(icon_path)
        if sizes:
            print(f"   New sizes: {sizes}")
            print(f"\n[OK] Icon should now display correctly at all sizes!")
        else:
            print("   Could not verify sizes")
    else:
        print("[ERROR] Icon file not found!")

    print("=" * 50)
