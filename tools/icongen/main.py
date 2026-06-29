#!/usr/bin/env python3
"""
SwiftRay Icon Generator

Generates cross-platform application icons from pre-sized source assets.
No processing is applied — assets are used as-is to preserve the original artwork.

Usage:
    python3 tools/icongen/main.py
    python3 tools/icongen/main.py --source /path/to/source.png
    python3 tools/icongen/main.py --sized-dir /path/to/sized/

Requires: Python 3.10+, Pillow
macOS .icns generation requires iconutil (Xcode Command Line Tools).
"""

import argparse
import io
import os
import struct
import subprocess
import sys
import tempfile
from pathlib import Path

try:
    from PIL import Image
except ImportError:
    print("Error: Pillow is required. Install with: pip install Pillow", file=sys.stderr)
    sys.exit(1)

SCRIPT_DIR = Path(__file__).resolve().parent
DEFAULT_SOURCE = str(SCRIPT_DIR / "source.png")
DEFAULT_SIZED_DIR = str(SCRIPT_DIR / "sized")
APP_NAME = "SwiftRay"

MACOS_ICONSET = [
    ("_16x16", 16),
    ("_16x16@2x", 32),
    ("_32x32", 32),
    ("_32x32@2x", 64),
    ("_128x128", 128),
    ("_128x128@2x", 256),
    ("_256x256", 256),
    ("_256x256@2x", 512),
    ("_512x512", 512),
    ("_512x512@2x", 1024),
]

WINDOWS_SIZES = [16, 24, 32, 48, 64, 128, 256]
LINUX_SIZES = [16, 24, 32, 48, 64, 128, 256, 512]


def validate_source(path: Path) -> Image.Image:
    if not path.exists():
        raise IconError(f"Source image not found: {path}")
    try:
        img = Image.open(str(path))
    except Exception as e:
        raise IconError(f"Cannot open image: {e}")
    if img.format != "PNG":
        raise IconError(f"Source must be PNG, got {img.format}")
    img = img.convert("RGBA")
    w, h = img.size
    if w != h:
        raise IconError(f"Source must be square. Got {w}x{h}.")
    if w < 256:
        raise IconError(f"Source too small: {w}px. Minimum 256px.")
    return img


class IconError(Exception):
    pass


def get_sized_image(sized_dir: Path, size: int) -> Image.Image | None:
    path = sized_dir / f"{size}.png"
    if path.exists():
        return Image.open(str(path)).convert("RGBA")
    return None


def get_best_image(sized_dir: Path, target_size: int, source: Image.Image) -> Image.Image:
    """Get the best available image for a target size."""
    # Try exact match from sized directory
    sized = get_sized_image(sized_dir, target_size)
    if sized is not None:
        return sized
    # Try double size and downscale
    for candidate in [target_size * 2, target_size * 4, 1024]:
        sized = get_sized_image(sized_dir, candidate)
        if sized is not None and sized.size[0] >= target_size:
            return sized.resize((target_size, target_size), Image.LANCZOS)
    # Fall back to source
    return source.resize((target_size, target_size), Image.LANCZOS)


def generate_macos_icns(sized_dir: Path, source: Image.Image, project_root: Path) -> Path | None:
    if subprocess.run(["which", "iconutil"], capture_output=True).returncode != 0:
        print("  -> iconutil not found, skipping .icns (generate on macOS)")
        return None

    icns_path = project_root / "build" / "bin" / f"{APP_NAME}.app" / "Contents" / "Resources" / "iconfile.icns"
    icns_path.parent.mkdir(parents=True, exist_ok=True)

    with tempfile.TemporaryDirectory() as tmpdir:
        iconset_dir = Path(tmpdir) / "AppIcon.iconset"
        iconset_dir.mkdir()

        for suffix, px in MACOS_ICONSET:
            img = get_best_image(sized_dir, px, source)
            img.save(str(iconset_dir / f"icon{suffix}.png"), "PNG", optimize=True)

        subprocess.run(
            ["iconutil", "-c", "icns", str(iconset_dir), "-o", str(icns_path)],
            check=True, capture_output=True,
        )

    print(f"  -> {icns_path.relative_to(project_root)}")
    return icns_path


def create_ico(images: list[tuple[Image.Image, int]], output_path: Path):
    num = len(images)
    image_data_list = []
    for img, size in images:
        buf = io.BytesIO()
        img.save(buf, "PNG")
        image_data_list.append((size, buf.getvalue()))

    header = struct.pack("<HHH", 0, 1, num)
    dir_entry_size = 16
    data_offset = 6 + dir_entry_size * num

    dir_entries = b""
    all_data = b""
    for size, data in image_data_list:
        w = size if size < 256 else 0
        h = size if size < 256 else 0
        entry = struct.pack("<BBBBHHII", w, h, 0, 0, 1, 32, len(data), data_offset + len(all_data))
        dir_entries += entry
        all_data += data

    with open(output_path, "wb") as f:
        f.write(header + dir_entries + all_data)


def generate_windows_ico(sized_dir: Path, source: Image.Image, project_root: Path) -> Path:
    ico_path = project_root / "build" / "windows" / "icon.ico"
    ico_path.parent.mkdir(parents=True, exist_ok=True)

    images = []
    for size in WINDOWS_SIZES:
        img = get_best_image(sized_dir, size, source)
        images.append((img, size))

    create_ico(images, ico_path)
    print(f"  -> {ico_path.relative_to(project_root)}")
    return ico_path


def generate_linux_pngs(sized_dir: Path, source: Image.Image, project_root: Path) -> Path:
    icons_dir = project_root / "build" / "icons"
    icons_dir.mkdir(parents=True, exist_ok=True)

    for size in LINUX_SIZES:
        img = get_best_image(sized_dir, size, source)
        out = icons_dir / f"icon-{size}x{size}.png"
        img.save(str(out), "PNG", optimize=True)

    print(f"  -> {icons_dir.relative_to(project_root)}/ ({len(LINUX_SIZES)} files)")
    return icons_dir


def generate_appicon(sized_dir: Path, source: Image.Image, project_root: Path) -> Path:
    img = get_best_image(sized_dir, 1024, source)
    out = project_root / "build" / "appicon.png"
    img.save(str(out), "PNG", optimize=True)
    print(f"  -> {out.relative_to(project_root)}")
    return out


def generate_sidebar_icon(sized_dir: Path, source: Image.Image, project_root: Path) -> Path:
    out_dir = project_root / "frontend" / "src" / "assets" / "images"
    out_dir.mkdir(parents=True, exist_ok=True)
    out = out_dir / "logo.png"

    img = get_best_image(sized_dir, 64, source)
    img.save(str(out), "PNG", optimize=True)
    print(f"  -> {out.relative_to(project_root)}")
    return out


def main():
    parser = argparse.ArgumentParser(description="Generate cross-platform app icons")
    parser.add_argument("--source", default=DEFAULT_SOURCE, help="Source PNG")
    parser.add_argument("--sized-dir", default=DEFAULT_SIZED_DIR, help="Pre-sized assets directory")
    parser.add_argument("--project-root", default=None, help="Project root")
    args = parser.parse_args()

    source_path = Path(args.source).expanduser()
    sized_dir = Path(args.sized_dir).expanduser()

    if args.project_root:
        project_root = Path(args.project_root)
    else:
        candidate = Path(__file__).resolve().parent.parent.parent
        project_root = candidate if (candidate / "wails.json").exists() else Path.cwd()

    if not (project_root / "wails.json").exists():
        print("Error: Cannot find wails.json. Specify --project-root.", file=sys.stderr)
        sys.exit(1)

    print("Validating source image...")
    source = validate_source(source_path)
    print(f"  Source: {source.size[0]}x{source.size[1]}, RGBA")
    sized_count = len(list(sized_dir.glob("*.png"))) if sized_dir.exists() else 0
    print(f"  Sized dir: {sized_dir} ({sized_count} files)")
    print(f"  Project: {project_root}")
    print()

    print("Generating icons (no processing applied)...")
    generate_appicon(sized_dir, source, project_root)
    generate_macos_icns(sized_dir, source, project_root)
    generate_windows_ico(sized_dir, source, project_root)
    generate_linux_pngs(sized_dir, source, project_root)
    generate_sidebar_icon(sized_dir, source, project_root)

    print()
    print("Done. All icons generated from source assets.")


if __name__ == "__main__":
    try:
        main()
    except IconError as e:
        print(f"Error: {e}", file=sys.stderr)
        sys.exit(1)
    except subprocess.CalledProcessError as e:
        print(f"Error: iconutil failed: {e}", file=sys.stderr)
        sys.exit(1)
