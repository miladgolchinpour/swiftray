package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

func BundledResourcesDir() string {
	if dir := findResourcesDir(); dir != "" {
		return dir
	}
	return defaultResourcesDir()
}

func defaultResourcesDir() string {
	exe, err := os.Executable()
	if err != nil {
		return ""
	}
	exeDir := filepath.Dir(exe)

	switch runtime.GOOS {
	case "darwin":
		return filepath.Join(exeDir, "..", "Resources")
	default:
		return filepath.Join(exeDir, "resources")
	}
}

func findResourcesDir() string {
	exe, err := os.Executable()
	if err != nil {
		return ""
	}
	exeDir := filepath.Dir(exe)
	if realExe, err := filepath.EvalSymlinks(exe); err == nil {
		exeDir = filepath.Dir(realExe)
	}

	candidates := []string{}

	switch runtime.GOOS {
	case "darwin":
		candidates = append(candidates,
			filepath.Join(exeDir, "..", "Resources"),
			filepath.Join(exeDir, "..", "Resources", "resources"),
			filepath.Join(exeDir, "resources"),
			filepath.Join(exeDir, "..", "MacOS", "resources"),
		)
	default:
		candidates = append(candidates,
			filepath.Join(exeDir, "resources"),
			filepath.Join(exeDir, "..", "resources"),
		)
	}

	if wd, err := os.Getwd(); err == nil {
		candidates = append(candidates, filepath.Join(wd, "resources"))
	}

	for _, c := range candidates {
		if dirHasGeoFiles(c) {
			return c
		}
	}

	return ""
}

func dirHasGeoFiles(dir string) bool {
	geoDir := filepath.Join(dir, "geo")
	_, errIP := os.Stat(filepath.Join(geoDir, "geoip.dat"))
	_, errSite := os.Stat(filepath.Join(geoDir, "geosite.dat"))
	return errIP == nil && errSite == nil
}

func BundledXrayDir() string {
	platform := PlatformName()
	return filepath.Join(BundledResourcesDir(), "xray", platform)
}

func BundledGeoDir() string {
	return filepath.Join(BundledResourcesDir(), "geo")
}

func BundledXrayPath() string {
	name := "xray"
	if runtime.GOOS == "windows" {
		name = "xray.exe"
	}
	return filepath.Join(BundledXrayDir(), name)
}

func BundledGeoIPPath() string {
	return filepath.Join(BundledGeoDir(), "geoip.dat")
}

func BundledGeoSitePath() string {
	return filepath.Join(BundledGeoDir(), "geosite.dat")
}

func ValidateBundledResources() error {
	xrayPath := BundledXrayPath()
	if _, err := os.Stat(xrayPath); err != nil {
		return fmt.Errorf("bundled xray binary not found at %s", xrayPath)
	}

	if runtime.GOOS != "windows" {
		info, err := os.Stat(xrayPath)
		if err != nil {
			return fmt.Errorf("cannot stat xray binary: %w", err)
		}
		if info.Mode().Perm()&0111 == 0 {
			return fmt.Errorf("xray binary is not executable (mode: %o)", info.Mode().Perm())
		}
	}

	if _, err := os.Stat(BundledGeoIPPath()); err != nil {
		return fmt.Errorf("bundled geoip.dat not found at %s", BundledGeoIPPath())
	}

	if _, err := os.Stat(BundledGeoSitePath()); err != nil {
		return fmt.Errorf("bundled geosite.dat not found at %s", BundledGeoSitePath())
	}

	return nil
}

type BundledResourceStatus struct {
	XrayExists   bool   `json:"xrayExists"`
	XrayPath     string `json:"xrayPath"`
	GeoIPExists  bool   `json:"geoIPExists"`
	GeoIPPath    string `json:"geoIPPath"`
	GeoSiteExists bool  `json:"geoSiteExists"`
	GeoSitePath  string `json:"geoSitePath"`
	Platform     string `json:"platform"`
	Valid        bool   `json:"valid"`
	Error        string `json:"error,omitempty"`
}

func GetBundledResourceStatus() BundledResourceStatus {
	status := BundledResourceStatus{
		XrayPath:     BundledXrayPath(),
		GeoIPPath:    BundledGeoIPPath(),
		GeoSitePath:  BundledGeoSitePath(),
		Platform:     PlatformName(),
	}

	if info, err := os.Stat(status.XrayPath); err == nil {
		status.XrayExists = info.Size() > 1024
	}
	if info, err := os.Stat(status.GeoIPPath); err == nil {
		status.GeoIPExists = info.Size() > 1024
	}
	if info, err := os.Stat(status.GeoSitePath); err == nil {
		status.GeoSiteExists = info.Size() > 1024
	}

	status.Valid = status.XrayExists && status.GeoIPExists && status.GeoSiteExists

	if err := ValidateBundledResources(); err != nil {
		status.Error = err.Error()
	}

	return status
}
