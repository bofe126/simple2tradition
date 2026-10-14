"""
regen_icos.py — regenerate Windows ICO files from PNG sources.

The old scripts/gen_icon.py draws an infinity-symbol design and writes
multi-size ICO files. After the new 简/繁 gradient design was added
to build/appicon.svg / appicon.png and the per-direction tray marks
were added to internal/tray/icon_{s2t,t2s}.png, those ICO files were
not refreshed — Windows tray surfaces and the NSIS installer icon
still showed the old ∞ design.

This script reads the current PNGs and emits multi-size ICO files at
{16, 32, 48, 64, 128, 256} for each (source → target) pair:

  build/appicon.png            → build/windows/icon.ico
                                → internal/tray/icon.ico
  internal/tray/icon_s2t.png   → internal/tray/icon_s2t.ico
  internal/tray/icon_t2s.png   → internal/tray/icon_t2s.ico

It also writes a 256x256 preview PNG next to each ICO for visual diffing.

Note: at 16x16 the 简/繁 text becomes illegible. If tray readability at
very small sizes matters more than brand consistency, a separate simple
mark for sizes ≤ 24px is the usual approach — not done here; the user
can layer that later if desired.

Usage:
  python scripts/regen_icos.py
"""

from __future__ import annotations

from pathlib import Path

from PIL import Image

PROJECT_ROOT = Path(__file__).resolve().parent.parent
# Coverage for Windows system-tray DPI buckets:
#   100%→16, 125%→20, 150%→24, 200%→32, 250%→40,
#   300%→48, plus 64/96/128/256 for taskbar / Explorer views.
# Each entry uses LANCZOS from the source 1024×1024 (or 512×512)
# PNG, so quality at every bucket is good as long as the source
# holds up. Note: at ≤24px the 简/繁 text in the design is
# effectively unreadable — a future "simplified mark" image for
# small sizes is the next iteration if tray legibility matters
# more than brand consistency.
SIZES = [16, 20, 24, 32, 40, 48, 64, 96, 128, 256]

# Auto-crop threshold. The source designs include a soft drop
# shadow (flood-opacity="0.25" ≈ max alpha 64) extending beyond
# the squircle. Anything below this alpha is treated as "edge
# padding" and trimmed off, so the icon fills its slot in the
# tray / installer instead of floating with a transparent margin.
# 128 cleanly separates the solid squircle (alpha=255) from the
# soft shadow (alpha ≤ ~64).
ALPHA_THRESHOLD = 128

# (source_png, target_ico) pairs.
PAIRS: list[tuple[Path, Path]] = [
    (
        PROJECT_ROOT / "build" / "appicon.png",
        PROJECT_ROOT / "build" / "windows" / "icon.ico",
    ),
    (
        PROJECT_ROOT / "build" / "appicon.png",
        PROJECT_ROOT / "internal" / "tray" / "icon.ico",
    ),
    (
        PROJECT_ROOT / "internal" / "tray" / "icon_s2t.png",
        PROJECT_ROOT / "internal" / "tray" / "icon_s2t.ico",
    ),
    (
        PROJECT_ROOT / "internal" / "tray" / "icon_t2s.png",
        PROJECT_ROOT / "internal" / "tray" / "icon_t2s.ico",
    ),
]


def make_ico(src: Image.Image, sizes: list[int]) -> dict[int, Image.Image]:
    """Render each requested size from the source PNG using LANCZOS."""
    return {s: src.resize((s, s), Image.LANCZOS) for s in sizes}


def autocrop(img: Image.Image) -> Image.Image:
    """Trim transparent edges so the icon fills its slot.

    Source designs (build/appicon.png, internal/tray/icon_*.png) carry
    the icon centered on a 1024×1024 (or 512×512) canvas with ~10%
    transparent margin on each side, plus a soft drop shadow that
    bleeds further. For system-tray use, that margin and shadow just
    waste pixels — at 16×16 each pixel costs ~6% of the slot.

    This crops to the visible bbox above ALPHA_THRESHOLD alpha, which
    keeps the solid squircle and excludes the soft shadow (max ~64
    alpha from a 0.25 flood-opacity filter).
    """
    alpha = img.split()[-1]
    mask = alpha.point(lambda a: 255 if a > ALPHA_THRESHOLD else 0)
    bbox = mask.getbbox()
    if not bbox:
        return img
    cropped = img.crop(bbox)
    if cropped.size[0] != cropped.size[1]:
        # Squircle designs are symmetric; this should never fire. If
        # it does, fall back to centring on a square canvas.
        side = max(cropped.size)
        square = Image.new("RGBA", (side, side), (0, 0, 0, 0))
        square.paste(
            cropped,
            ((side - cropped.size[0]) // 2, (side - cropped.size[1]) // 2),
            cropped,
        )
        return square
    return cropped


def write_ico(images: dict[int, Image.Image], out_path: Path) -> None:
    """Save a multi-size ICO. Pillow picks the entries it stores.

    Largest goes first as the base; the rest are appended.
    """
    out_path.parent.mkdir(parents=True, exist_ok=True)
    sizes_sorted = sorted(images.keys(), reverse=True)
    base_img = images[sizes_sorted[0]]
    append = [images[s] for s in sizes_sorted[1:]]
    sizes_arg = [(s, s) for s in sorted(images.keys())]
    base_img.save(
        out_path,
        format="ICO",
        sizes=sizes_arg,
        append_images=append,
    )


def main() -> None:
    # Cache images per source so we only decode each PNG once.
    cache: dict[Path, dict[int, Image.Image]] = {}

    for src_path, ico_path in PAIRS:
        if not src_path.is_file():
            print(f"[SKIP] source missing: {src_path.relative_to(PROJECT_ROOT)}")
            continue

        if src_path not in cache:
            src = Image.open(src_path)
            if src.mode != "RGBA":
                src = src.convert("RGBA")
            cropped = autocrop(src)
            crop_note = ""
            if cropped.size != src.size:
                crop_note = (
                    f"  autocrop {src.size}→{cropped.size} "
                    f"(+{((cropped.size[0]/src.size[0])*100 - 100):+.1f}%)"
                )
            cache[src_path] = make_ico(cropped, SIZES)
            print(f"Source: {src_path.relative_to(PROJECT_ROOT)} "
                  f"({src.size} → cropped {cropped.size}){crop_note}")

        images = cache[src_path]
        write_ico(images, ico_path)
        preview = ico_path.with_name(ico_path.stem + ".preview.png")
        images[256].save(preview, format="PNG")
        print(f"[OK] {ico_path.relative_to(PROJECT_ROOT)} "
              f"(sizes: {SIZES}, preview: {preview.name})")


if __name__ == "__main__":
    main()