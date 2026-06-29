# SwiftRay Resource Manager

Standalone tool for managing bundled runtime assets (Xray binary, geo files).

## Design Principle

The application never downloads anything at runtime. All required assets are bundled during the release process. This tool manages those assets.

## Quick Start

```bash
# Download latest Xray for current platform
go run ./tools/resmanager fetch xray

# Download latest Xray for all platforms
go run ./tools/resmanager fetch xray --all

# Download latest geo files
go run ./tools/resmanager fetch geo

# Verify all resources
go run ./tools/resmanager verify

# Show bundled versions
go run ./tools/resmanager list

# Clean all resources
go run ./tools/resmanager clean
```

## Commands

### `fetch xray`

Download Xray binary for specified platform(s).

```bash
# Latest version, current platform
go run ./tools/resmanager fetch xray

# Specific version, current platform
go run ./tools/resmanager fetch xray --version v1.8.4

# Latest version, specific platform
go run ./tools/resmanager fetch xray --platform darwin-arm64-v8a

# Latest version, all platforms
go run ./tools/resmanager fetch xray --all
```

**Platforms:**
- `darwin-arm64-v8a` — macOS Apple Silicon
- `darwin-amd64` — macOS Intel
- `windows-amd64` — Windows x64
- `linux-amd64` — Linux x64
- `linux-arm64` — Linux ARM64

### `fetch geo`

Download latest geoip.dat and geosite.dat.

```bash
go run ./tools/resmanager fetch geo
```

### `verify`

Verify all bundled resources exist and are valid.

```bash
go run ./tools/resmanager verify
```

Checks:
- All geo files exist
- All Xray binaries exist
- Unix binaries are executable
- VERSION files are present

### `list`

Show bundled resource versions and sizes.

```bash
go run ./tools/resmanager list
```

### `clean`

Remove all bundled resources.

```bash
go run ./tools/resmanager clean
```

## Options

- `--verbose`, `-v` — Show detailed output
- `--dry-run` — Show what would be done without making changes
- `--help`, `-h` — Show help

## Directory Structure

After running fetch commands:

```
resources/
├── xray/
│   ├── darwin-arm64-v8a/
│   │   ├── xray
│   │   └── VERSION
│   ├── darwin-amd64/
│   │   ├── xray
│   │   └── VERSION
│   ├── windows-amd64/
│   │   ├── xray.exe
│   │   └── VERSION
│   ├── linux-amd64/
│   │   ├── xray
│   │   └── VERSION
│   └── linux-arm64/
│       ├── xray
│       └── VERSION
└── geo/
    ├── geoip.dat
    ├── geosite.dat
    └── VERSION
```

## Release Workflow

### 1. Update Xray version

```bash
# Download new version for all platforms
go run ./tools/resmanager fetch xray --all --version v1.8.5

# Verify
go run ./tools/resmanager verify
go run ./tools/resmanager list
```

### 2. Update geo files

```bash
# Download latest
go run ./tools/resmanager fetch geo

# Verify
go run ./tools/resmanager verify
```

### 3. Verify before release

```bash
go run ./tools/resmanager verify
go run ./tools/resmanager list
```

### 4. Build application

```bash
wails build
```

The build will include all files from `resources/` in the application bundle.

## Updating Xray

1. Check latest release: https://github.com/XTLS/Xray-core/releases
2. Run: `go run ./tools/resmanager fetch xray --all --version <version>`
3. Run: `go run ./tools/resmanager verify`
4. Commit the updated resources
5. Build the application

## Updating Geo Files

1. Run: `go run ./tools/resmanager fetch geo`
2. Run: `go run ./tools/resmanager verify`
3. Commit the updated files
4. Build the application

## Troubleshooting

### "bundled xray binary not found"

Run: `go run ./tools/resmanager fetch xray --all`

### "xray binary is not executable"

Run: `chmod +x resources/xray/<platform>/xray`

### "geoip.dat not found"

Run: `go run ./tools/resmanager fetch geo`

## Architecture

```
tools/resmanager/          ← This tool (standalone)
    main.go

resources/                 ← Bundled assets (source-controlled)
    xray/
    geo/

app/utils/resources.go     ← Application code (reads resources)
app/services/xray.go       ← Application code (uses resources)
```

The application only reads from `resources/`. It never writes to it, never downloads, and never knows how the resources were obtained.
