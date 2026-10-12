"""
gen_icon.py — render the simple2tradition app/tray icon.

Concept: a black infinity (∞) symbol on a pure white background. The
shape is built from two overlapping black "donuts" (large circle minus
smaller inner circle), one on each side, meeting at the vertical
centerline. This reads as "transformation / loop" without using text.

Sizing strategy: render at 8x the target size with anti-aliased
geometry, then downsample with LANCZOS. This keeps the 16x16 tray
icon readable (no aliasing artifacts) while the 256x256 stays crisp.

Output:
  - icon.ico   multi-size ICO with 16, 32, 48, 64, 128, 256
  - preview.png  256x256 PNG for human inspection

Usage:
  python scripts/gen_icon.py            # writes next to this file
  python scripts/gen_icon.py --out X    # writes to directory X
"""

from __future__ import annotations

import argparse
import math
import os
from pathlib import Path

from PIL import Image, ImageDraw

# Geometry tuned so the icon reads well even at 16x16:
#   - outer radius covers ~64% of canvas width (each lobe)
#   - inner radius is ~38% of canvas width (donut hole)
#   - centers offset by 24% so the lobes overlap at the middle
# The values are fractions of the canvas size.
OUTER_RADIUS_FRAC = 0.32
INNER_RADIUS_FRAC = 0.18
CENTER_OFFSET_FRAC = 0.24
PADDING_FRAC = 0.04  # keep some breathing room near the edges


def make_infinity(size: int) -> Image.Image:
    """Render the infinity icon at the given pixel size."""
    scale = 8  # supersample for anti-aliasing
    big = max(size * scale, 32)
    img = Image.new("RGBA", (big, big), (255, 255, 255, 255))
    draw = ImageDraw.Draw(img)

    cx = big / 2.0
    cy = big / 2.0
    r_out = big * OUTER_RADIUS_FRAC
    r_in = big * INNER_RADIUS_FRAC
    dx = big * CENTER_OFFSET_FRAC

    # Two filled black circles (the lobes)
    for sign in (-1, 1):
        cxi = cx + sign * dx
        draw.ellipse(
            [cxi - r_out, cy - r_out, cxi + r_out, cy + r_out],
            fill=(0, 0, 0, 255),
        )

    # Two white circles punched out of the middle of each lobe,
    # forming the donut hole. The two donuts touch at the center.
    for sign in (-1, 1):
        cxi = cx + sign * dx
        draw.ellipse(
            [cxi - r_in, cy - r_in, cxi + r_in, cy + r_in],
            fill=(255, 255, 255, 255),
        )

    # LANCZOS downsample for the final size. The result is kept as
    # RGBA so the .ico writer can preserve the white background.
    return img.resize((size, size), Image.LANCZOS)


def write_ico(images: dict[int, Image.Image], out_path: Path) -> None:
    """Save a multi-size .ico file. Pillow picks the sizes it stores.

    Pillow's ICO writer accepts a list of (size, image) tuples via
    `sizes=` plus `append_images=`. We order the list from largest
    (best quality) to smallest.
    """
    sizes_sorted = sorted(images.keys(), reverse=True)
    # The first image acts as the "base"; subsequent sizes are
    # appended. Pillow resizes the base down for any missing sizes,
    # so we include explicit entries for all the sizes we want.
    base_size = sizes_sorted[0]
    base_img = images[base_size]
    append = [images[s] for s in sizes_sorted[1:]]
    sizes_arg = [(s, s) for s in sorted(images.keys())]
    base_img.save(
        out_path,
        format="ICO",
        sizes=sizes_arg,
        append_images=append,
    )


def main() -> None:
    ap = argparse.ArgumentParser(description=__doc__.splitlines()[0])
    ap.add_argument(
        "--out",
        default=str(Path(__file__).parent),
        help="Output directory (default: alongside this script).",
    )
    args = ap.parse_args()

    out_dir = Path(args.out)
    out_dir.mkdir(parents=True, exist_ok=True)

    sizes = [16, 32, 48, 64, 128, 256]
    images: dict[int, Image.Image] = {s: make_infinity(s) for s in sizes}

    ico_path = out_dir / "icon.ico"
    png_path = out_dir / "preview.png"

    write_ico(images, ico_path)
    images[256].save(png_path, format="PNG")

    print(f"[OK] Wrote {ico_path}  (sizes: {sizes})")
    print(f"[OK] Wrote {png_path}  (256x256 preview)")


if __name__ == "__main__":
    main()
