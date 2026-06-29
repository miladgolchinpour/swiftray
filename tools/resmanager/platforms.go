package main

import (
	"fmt"
	"runtime"
)

type Platform struct {
	Name   string
	GOOS   string
	GOARCH string
	Zip    string
}

var platforms = []Platform{
	{"darwin-arm64-v8a", "darwin", "arm64", "Xray-macos-arm64-v8a.zip"},
	{"darwin-amd64", "darwin", "amd64", "Xray-macos-64.zip"},
	{"windows-amd64", "windows", "amd64", "Xray-windows-64.zip"},
	{"linux-amd64", "linux", "amd64", "Xray-linux-64.zip"},
	{"linux-arm64", "linux", "arm64", "Xray-linux-arm64-v8a.zip"},
}

func CurrentPlatform() Platform {
	for _, p := range platforms {
		if p.GOOS == runtime.GOOS && p.GOARCH == runtime.GOARCH {
			return p
		}
	}
	return Platform{Name: "unknown", GOOS: runtime.GOOS, GOARCH: runtime.GOARCH}
}

func FindPlatform(name string) (Platform, bool) {
	for _, p := range platforms {
		if p.Name == name {
			return p, true
		}
	}
	return Platform{}, false
}

func ListPlatformNames() []string {
	names := make([]string, len(platforms))
	for i, p := range platforms {
		names[i] = p.Name
	}
	return names
}

func (p Platform) XrayBinaryName() string {
	if p.GOOS == "windows" {
		return "xray.exe"
	}
	return "xray"
}

func (p Platform) String() string {
	return fmt.Sprintf("%s (%s/%s)", p.Name, p.GOOS, p.GOARCH)
}
