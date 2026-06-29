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

## Installation

Download the latest release from the Releases page.

| Platform | Package |
|----------|---------|
| macOS Apple Silicon | `.dmg` |
| macOS Intel | `.dmg` |
| Windows | Portable ZIP |
| Linux | `.tar.gz` |

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

There is also a Swift version which provides native macOS (Apple Silicon) experience.

## License

MIT
