package utils

import (
	"os"
	"path/filepath"
	"runtime"
)

func AppDataDir() string {
	switch runtime.GOOS {
	case "darwin":
		home, _ := os.UserHomeDir()
		return filepath.Join(home, "Library", "Application Support", "SwiftRay")
	case "windows":
		appData := os.Getenv("APPDATA")
		if appData == "" {
			home, _ := os.UserHomeDir()
			appData = filepath.Join(home, "AppData", "Roaming")
		}
		return filepath.Join(appData, "SwiftRay")
	default:
		home, _ := os.UserHomeDir()
		return filepath.Join(home, ".config", "swiftray")
	}
}

func ConfigPath() string {
	return filepath.Join(AppDataDir(), "config.json")
}

func PlatformName() string {
	switch runtime.GOOS {
	case "darwin":
		if runtime.GOARCH == "arm64" {
			return "darwin-arm64-v8a"
		}
		return "darwin-amd64"
	case "windows":
		return "windows-amd64"
	case "linux":
		if runtime.GOARCH == "arm64" {
			return "linux-arm64"
		}
		return "linux-amd64"
	default:
		return ""
	}
}
