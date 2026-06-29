# SwiftRay

SwiftRay is a cross-platform GUI for Xray-core built with Go, Wails, Vue 3 and TypeScript.

It focuses on being lightweight, fast and easy to use while providing a simple desktop experience on macOS, Windows and Linux.

## Features

- Built-in Xray runtime
- Subscription management
- Local node management
- URL latency testing
- System proxy management
- Live Xray logs
- IP information
- Automatic settings persistence
- Xray runtime updater
- Cross-platform support

## Platforms

- macOS (Apple Silicon)
- macOS (Intel)
- Windows x64
- Linux x64

## Download & Installation

| Platform | Download |
|----------|---------|
| macOS Apple Silicon | [SwiftRay-macos-arm64.dmg](https://github.com/MiladGolchinpour/SwiftRay/releases/download/v0.1.0/SwiftRay-macos-arm64.dmg) |
| macOS Intel | [SwiftRay-macos-amd64.dmg](https://github.com/MiladGolchinpour/SwiftRay/releases/download/v0.1.0/SwiftRay-macos-amd64.dmg) |
| Windows x64 (Portable ZIP) | [SwiftRay-windows-amd64.zip](https://github.com/MiladGolchinpour/SwiftRay/releases/download/v0.1.0/SwiftRay-windows-amd64.zip) |
| Linux x64 | [SwiftRay-linux-amd64.tar.gz](https://github.com/MiladGolchinpour/SwiftRay/releases/download/v0.1.0/SwiftRay-linux-amd64.tar.gz) |

### macOS

- **DMG (.dmg)** — Open the disk image and drag **SwiftRay.app** into the **Applications** folder.

If Gatekeeper blocks the app, open **Terminal** and run:

```bash

xattr -dr com.apple.quarantine /Applications/SwiftRay.app

```

### Windows

- **Portable (.zip)** — Extract the archive and run `SwiftRay.exe`.

If Windows SmartScreen appears, click **More info → Run anyway**. This can happen because the application is not code signed.

### Linux

- **Archive (.tar.gz)** — Extract the archive, make the binary executable, and run it.

```bash

chmod +x SwiftRay

./SwiftRay

```

Install the required WebKitGTK runtime if it is not already available on your distribution.

## Build

Requirements:

- Go 1.23+
- Node.js 20+
- Wails v2

Fetch bundled resources:

```bash
go run ./tools/resmanager fetch xray --all
```

Build:

```bash
wails build
```

## Project

```
app/          Go backend
frontend/     Vue 3 frontend
resources/    Bundled Xray runtime
tools/        Resource manager & utilities
```

## License

MIT
