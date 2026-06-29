# SwiftRay Icon Generator

Generates cross-platform application icons from a single high-resolution source PNG.

## Principle

The source image is the **only editable asset**. Every icon asset is generated deterministically — no manual editing of output files.

```
source.png  ──►  appicon.png (macOS .icns, Windows .ico, Linux PNGs, frontend logo)
```

## Requirements

- Python 3.10+
- Pillow (`pip install Pillow`)
- macOS: Xcode Command Line Tools (for `iconutil`)

## Usage

```bash
# Default source (~/Desktop/appiconraw.png)
make icons

# Custom source
make icons-from SRC=/path/to/source.png

# Direct invocation
python3 tools/icongen/main.py
python3 tools/icongen/main.py --source /path/to/source.png
python3 tools/icongen/main.py --project-root /path/to/project
```

## Source Requirements

| Requirement | Minimum |
|---|---|
| Format | PNG |
| Dimensions | 256×256 px (recommended: 1024×1024+) |
| Aspect ratio | Square (within 1px tolerance) |
| Color mode | RGBA (RGB accepted with warning) |

## Generated Assets

| Asset | Path | Purpose |
|---|---|---|
| appicon.png | `build/appicon.png` | Wails build icon (1024×1024) |
| iconfile.icns | `build/bin/SwiftRay.app/Contents/Resources/` | macOS application icon |
| icon.ico | `build/windows/icon.ico` | Windows application icon (7 sizes) |
| icon-*.png | `build/icons/` | Linux icon set (8 sizes: 16–512) |
| logo.png | `frontend/src/assets/images/` | Frontend sidebar logo (64×64) |

## Processing

- **Rounded corners**: 18% radius, 4× supersampled anti-aliasing
- **Padding**: 5% safe area around content
- **Resampling**: LANCZOS (highest quality downscaling)
- **Format**: RGBA throughout, PNG output with optimization

## Validation

The tool fails with a clear error if:

- Source file does not exist
- Source is not a PNG file
- Source is smaller than 256×256
- Source is not square
- Pillow or iconutil is unavailable

## Integration

- **Makefile**: `make icons` / `make icons-from SRC=...`
- **CI/CD**: `release.yml` runs icon generation before every build
- **Wails**: Automatically picks up `build/appicon.png` for all platforms

## Regenerating Icons

When the branding changes, update `~/Desktop/appiconraw.png` and run:

```bash
make icons
```

All downstream assets are regenerated automatically.
