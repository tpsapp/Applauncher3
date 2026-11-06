#!/usr/bin/env python3
"""
scripts/generate_icons.py

Generates cross-platform application branding icon assets for AppLauncher3:
  - build/appicon.png (512x512 Master RGBA PNG for Linux GTK window icon)
  - build/windows/icon.ico (Multi-resolution Windows application icon)
  - frontend/public/icon.png (Webview favicon asset)
  - frontend/public/favicon.ico (Multi-resolution web fallback)

The icon matches the in-app header badge:
  - Squircle badge with rounded corners
  - Gradient: indigo-600 (#4f46e5) -> indigo-500 (#6366f1) -> cyan-400 (#22d3ee)
  - Centered 🚀 rocket emblem
"""

import os
import shutil
import subprocess
import sys
import tempfile
from PIL import Image, ImageDraw, ImageFont

PROJECT_ROOT = os.path.abspath(os.path.join(os.path.dirname(__file__), ".."))
BUILD_DIR = os.path.join(PROJECT_ROOT, "build")
BUILD_WINDOWS_DIR = os.path.join(BUILD_DIR, "windows")
PUBLIC_DIR = os.path.join(PROJECT_ROOT, "frontend", "public")

HTML_TEMPLATE = """<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<style>
  * { margin: 0; padding: 0; box-sizing: border-box; }
  body {
    width: 512px;
    height: 512px;
    background: transparent;
    display: flex;
    align-items: center;
    justify-content: center;
  }
  .icon-box {
    width: 472px;
    height: 472px;
    border-radius: 118px;
    background: linear-gradient(135deg, #22d3ee 0%, #6366f1 50%, #4f46e5 100%);
    box-shadow: 0 16px 36px rgba(79, 70, 229, 0.45);
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 260px;
    line-height: 1;
    user-select: none;
  }
</style>
</head>
<body>
  <div class="icon-box">🚀</div>
</body>
</html>
"""


def render_with_chrome(output_png_path: str) -> bool:
    """Renders pixel-perfect 512x512 icon using headless Chrome/Chromium if available."""
    chrome_bins = [
        "google-chrome-stable",
        "google-chrome",
        "chromium",
        "chromium-browser",
    ]
    selected_bin = None
    for b in chrome_bins:
        path = shutil.which(b)
        if path:
            selected_bin = path
            break

    if not selected_bin:
        return False

    with tempfile.NamedTemporaryFile("w", suffix=".html", delete=False) as f:
        f.write(HTML_TEMPLATE)
        temp_html = f.name

    try:
        cmd = [
            selected_bin,
            "--headless",
            "--disable-gpu",
            "--force-device-scale-factor=1",
            "--window-size=512,512",
            "--default-background-color=00000000",
            f"--screenshot={output_png_path}",
            temp_html,
        ]
        res = subprocess.run(cmd, capture_output=True, text=True, timeout=15)
        return res.returncode == 0 and os.path.exists(output_png_path)
    except Exception as e:
        print(f"Warning: Chrome rendering failed: {e}", file=sys.stderr)
        return False
    finally:
        if os.path.exists(temp_html):
            os.remove(temp_html)


def render_with_pillow(output_png_path: str):
    """Fallback generator using Pillow if headless browser is unavailable."""
    width, height = 512, 512
    im = Image.new("RGBA", (width, height), (0, 0, 0, 0))

    # Create gradient squircle
    mask = Image.new("L", (width, height), 0)
    draw_mask = ImageDraw.Draw(mask)
    draw_mask.rounded_rectangle([20, 20, 492, 492], radius=118, fill=255)

    gradient = Image.new("RGBA", (width, height))
    draw_grad = ImageDraw.Draw(gradient)
    c1 = (34, 211, 238)   # cyan-400
    c2 = (99, 102, 241)   # indigo-500
    c3 = (79, 70, 229)   # indigo-600

    for y in range(height):
        factor = y / height
        if factor < 0.5:
            sub = factor * 2.0
            r = int(c1[0] + (c2[0] - c1[0]) * sub)
            g = int(c1[1] + (c2[1] - c1[1]) * sub)
            b = int(c1[2] + (c2[2] - c1[2]) * sub)
        else:
            sub = (factor - 0.5) * 2.0
            r = int(c2[0] + (c3[0] - c2[0]) * sub)
            g = int(c2[1] + (c3[1] - c2[1]) * sub)
            b = int(c2[2] + (c3[2] - c2[2]) * sub)
        draw_grad.line([(0, y), (width, y)], fill=(r, g, b, 255))

    im.paste(gradient, (0, 0), mask)

    # Attempt Noto Color Emoji font
    font_path = "/usr/share/fonts/noto/NotoColorEmoji.ttf"
    if os.path.exists(font_path):
        try:
            font = ImageFont.truetype(font_path, 109)
            emoji_layer = Image.new("RGBA", (150, 150), (0, 0, 0, 0))
            draw_emoji = ImageDraw.Draw(emoji_layer)
            draw_emoji.text((10, 10), "🚀", font=font, embedded_color=True)
            # Scale up to ~270px
            scaled_emoji = emoji_layer.resize((270, 270), Image.Resampling.LANCZOS)
            offset_x = (width - 270) // 2
            offset_y = (height - 270) // 2
            im.paste(scaled_emoji, (offset_x, offset_y), scaled_emoji)
        except Exception:
            pass

    im.save(output_png_path, "PNG")


def main():
    os.makedirs(BUILD_DIR, exist_ok=True)
    os.makedirs(BUILD_WINDOWS_DIR, exist_ok=True)
    os.makedirs(PUBLIC_DIR, exist_ok=True)

    master_png = os.path.join(BUILD_DIR, "appicon.png")

    print("[1/4] Generating master PNG icon (512x512)...")
    success = render_with_chrome(master_png)
    if not success or not os.path.exists(master_png):
        print("Falling back to Pillow renderer...")
        render_with_pillow(master_png)

    # Verify and load master image
    img = Image.open(master_png).convert("RGBA")
    if img.size != (512, 512):
        img = img.resize((512, 512), Image.Resampling.LANCZOS)
        img.save(master_png, "PNG")

    print(f"  ✓ Saved master PNG: {master_png} ({img.size[0]}x{img.size[1]})")

    # Generate Windows multi-resolution ICO
    print("[2/4] Generating Windows multi-resolution icon (icon.ico)...")
    win_ico = os.path.join(BUILD_WINDOWS_DIR, "icon.ico")
    ico_sizes = [(16, 16), (32, 32), (48, 48), (64, 64), (128, 128), (256, 256)]
    img.save(win_ico, format="ICO", sizes=ico_sizes)
    print(f"  ✓ Saved Windows ICO: {win_ico} with {len(ico_sizes)} resolutions")

    # Generate Frontend webview favicon assets
    print("[3/4] Exporting frontend webview icon assets...")
    public_png = os.path.join(PUBLIC_DIR, "icon.png")
    shutil.copyfile(master_png, public_png)
    print(f"  ✓ Copied web favicon: {public_png}")

    public_ico = os.path.join(PUBLIC_DIR, "favicon.ico")
    img.save(public_ico, format="ICO", sizes=[(16, 16), (32, 32), (48, 48)])
    print(f"  ✓ Saved web favicon.ico: {public_ico}")

    print("[4/4] Asset generation complete!")


if __name__ == "__main__":
    main()
