package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const (
	xrayRepo = "XTLS/Xray-core"
	githubAPI = "https://api.github.com/repos"
)

var (
	verbose bool
	dryRun  bool
	cfg     Config
)

func main() {
	cfg = LoadConfig()

	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	for _, arg := range os.Args[2:] {
		switch arg {
		case "--verbose", "-v":
			verbose = true
		case "--dry-run":
			dryRun = true
		}
	}

	switch os.Args[1] {
	case "fetch":
		cmdFetch()
	case "verify":
		cmdVerify()
	case "list":
		cmdList()
	case "clean":
		cmdClean()
	case "help", "--help", "-h":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`SwiftRay Resource Manager

Usage:
  resmanager <command> [options]

Commands:
  fetch xray [--version <ver>] [--platform <plat>] [--all]
      Download Xray binary. Default: latest version, current platform.

  fetch geo
      Download latest geoip.dat and geosite.dat.

  verify
      Verify all bundled resources exist and are valid.

  list
      Show bundled resource versions and status.

  clean
      Remove all bundled resources.

Options:
  --verbose, -v    Show detailed output.
  --dry-run        Show what would be done without making changes.
  --help, -h       Show this help message.

Configuration (.env.resmanager or .env):
  XRAY_VERSION     Xray version (default: latest)
  XRAY_MIRROR      Download mirror URL
  GEO_IP_URL       Custom geoip.dat URL
  GEO_SITE_URL     Custom geosite.dat URL
  RESOURCES_DIR    Output directory (default: resources)`)
}

func cmdFetch() {
	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "Usage: resmanager fetch <xray|geo> [options]")
		os.Exit(1)
	}

	switch os.Args[2] {
	case "xray":
		cmdFetchXray()
	case "geo":
		cmdFetchGeo()
	default:
		fmt.Fprintf(os.Stderr, "Unknown fetch target: %s\n", os.Args[2])
		os.Exit(1)
	}
}

func cmdFetchXray() {
	version := cfg.XrayVersion
	platformName := ""
	all := false

	for i := 3; i < len(os.Args); i++ {
		switch os.Args[i] {
		case "--version", "-V":
			if i+1 < len(os.Args) {
				version = os.Args[i+1]
				i++
			}
		case "--platform", "-p":
			if i+1 < len(os.Args) {
				platformName = os.Args[i+1]
				i++
			}
		case "--all":
			all = true
		case "--verbose", "-v":
			verbose = true
		case "--dry-run":
			dryRun = true
		}
	}

	if version == "latest" {
		fmt.Println("Fetching latest Xray version...")
		latestVersion, err := getLatestVersion()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to get latest version: %v\n", err)
			os.Exit(1)
		}
		version = latestVersion
		fmt.Printf("Latest version: %s\n", version)
	}

	targets := platforms
	if platformName != "" && !all {
		p, found := FindPlatform(platformName)
		if !found {
			fmt.Fprintf(os.Stderr, "Unknown platform: %s\nAvailable: %s\n", platformName, strings.Join(ListPlatformNames(), ", "))
			os.Exit(1)
		}
		targets = []Platform{p}
	}

	for _, p := range targets {
		fmt.Printf("\n--- %s ---\n", p.Name)
		if err := downloadXray(version, p); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	}

	fmt.Println("\nDone!")
}

func downloadXray(version string, p Platform) error {
	url := cfg.XrayZipURL(p)
	destDir := cfg.ResourcePath("xray", p.Name)
	zipPath := filepath.Join(destDir, "xray.zip")

	if dryRun {
		fmt.Printf("[dry-run] Would download %s to %s\n", url, destDir)
		return nil
	}

	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return fmt.Errorf("create dir: %w", err)
	}

	fmt.Printf("Downloading %s...\n", p.Zip)
	if err := downloadFile(url, zipPath); err != nil {
		return fmt.Errorf("download: %w", err)
	}

	fmt.Println("Extracting...")
	if err := extractZip(zipPath, destDir); err != nil {
		return fmt.Errorf("extract: %w", err)
	}

	xrayName := p.XrayBinaryName()
	extracted := findExtractedBinary(destDir, xrayName)
	if extracted == "" {
		return fmt.Errorf("xray binary not found in zip")
	}

	finalPath := filepath.Join(destDir, xrayName)
	if extracted != finalPath {
		os.Remove(finalPath)
		if err := os.Rename(extracted, finalPath); err != nil {
			return fmt.Errorf("move binary: %w", err)
		}
	}

	os.Remove(zipPath)
	cleanupNestedDirs(destDir)

	if p.GOOS != "windows" {
		if err := os.Chmod(finalPath, 0o755); err != nil {
			return fmt.Errorf("chmod: %w", err)
		}
	}

	versionFile := filepath.Join(destDir, "VERSION")
	os.WriteFile(versionFile, []byte(version), 0o644)

	fmt.Printf("Installed: %s (%s)\n", finalPath, version)
	return nil
}

func findExtractedBinary(dir, name string) string {
	// Check direct
	if _, err := os.Stat(filepath.Join(dir, name)); err == nil {
		return filepath.Join(dir, name)
	}
	// Check nested directories
	entries, _ := os.ReadDir(dir)
	for _, e := range entries {
		if e.IsDir() {
			if _, err := os.Stat(filepath.Join(dir, e.Name(), name)); err == nil {
				return filepath.Join(dir, e.Name(), name)
			}
		}
	}
	return ""
}

func cleanupNestedDirs(dir string) {
	entries, _ := os.ReadDir(dir)
	for _, e := range entries {
		if e.IsDir() {
			os.RemoveAll(filepath.Join(dir, e.Name()))
		}
	}
}

func cmdFetchGeo() {
	destDir := cfg.ResourcePath("geo")
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "Error: create dir: %v\n", err)
		os.Exit(1)
	}

	type geoFile struct {
		name string
		url  string
	}

	files := []geoFile{
		{"geoip.dat", cfg.GeoIPURL},
		{"geosite.dat", cfg.GeoSiteURL},
	}

	for _, f := range files {
		if dryRun {
			fmt.Printf("[dry-run] Would download %s to %s\n", f.url, filepath.Join(destDir, f.name))
			continue
		}

		fmt.Printf("Downloading %s...\n", f.name)
		if err := downloadFile(f.url, filepath.Join(destDir, f.name)); err != nil {
			fmt.Fprintf(os.Stderr, "Error downloading %s: %v\n", f.name, err)
			os.Exit(1)
		}
		fmt.Printf("Installed: %s\n", filepath.Join(destDir, f.name))
	}

	versionFile := filepath.Join(destDir, "VERSION")
	os.WriteFile(versionFile, []byte(time.Now().Format(time.RFC3339)), 0o644)

	fmt.Println("\nDone!")
}

func cmdVerify() {
	fmt.Println("Verifying bundled resources...\n")
	allValid := true

	geoDir := cfg.ResourcePath("geo")
	for _, f := range []string{"geoip.dat", "geosite.dat"} {
		path := filepath.Join(geoDir, f)
		if info, err := os.Stat(path); err != nil {
			fmt.Printf("  MISSING  %s\n", path)
			allValid = false
		} else {
			fmt.Printf("  OK       %s (%d bytes)\n", path, info.Size())
		}
	}

	for _, p := range platforms {
		xrayName := p.XrayBinaryName()
		path := cfg.ResourcePath("xray", p.Name, xrayName)
		versionPath := cfg.ResourcePath("xray", p.Name, "VERSION")

		if info, err := os.Stat(path); err != nil {
			fmt.Printf("  MISSING  %s (%s)\n", path, p.Name)
			allValid = false
		} else {
			version := "?"
			if v, err := os.ReadFile(versionPath); err == nil {
				version = strings.TrimSpace(string(v))
			}

			executable := true
			if p.GOOS != "windows" {
				executable = info.Mode().Perm()&0111 != 0
			}

			status := "OK"
			if !executable {
				status = "NOT EXECUTABLE"
				allValid = false
			}

			fmt.Printf("  %-8s %s (%s, %d bytes)\n", status, path, version, info.Size())
		}
	}

	fmt.Println()
	if allValid {
		fmt.Println("All resources are valid.")
	} else {
		fmt.Println("Some resources are missing or invalid.")
		os.Exit(1)
	}
}

func cmdList() {
	fmt.Println("Bundled Resources\n")

	geoDir := cfg.ResourcePath("geo")
	fmt.Println("Geo files:")
	for _, f := range []string{"geoip.dat", "geosite.dat"} {
		path := filepath.Join(geoDir, f)
		if info, err := os.Stat(path); err == nil {
			fmt.Printf("  %s  %d bytes  %s\n", f, info.Size(), info.ModTime().Format("2006-01-02"))
		} else {
			fmt.Printf("  %s  (not found)\n", f)
		}
	}

	fmt.Println("\nXray binaries:")
	for _, p := range platforms {
		xrayName := p.XrayBinaryName()
		path := cfg.ResourcePath("xray", p.Name, xrayName)
		versionPath := cfg.ResourcePath("xray", p.Name, "VERSION")

		version := "unknown"
		if v, err := os.ReadFile(versionPath); err == nil {
			version = strings.TrimSpace(string(v))
		}

		if info, err := os.Stat(path); err == nil {
			fmt.Printf("  %-20s  v%-10s  %d bytes\n", p.Name, version, info.Size())
		} else {
			fmt.Printf("  %-20s  (not found)\n", p.Name)
		}
	}
}

func cmdClean() {
	if dryRun {
		fmt.Printf("[dry-run] Would remove %s/\n", cfg.OutputDir)
		return
	}

	fmt.Printf("Removing %s/...\n", cfg.OutputDir)
	if err := os.RemoveAll(cfg.OutputDir); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Done!")
}

// --- Helpers ---

func getLatestVersion() (string, error) {
	url := fmt.Sprintf("%s/%s/releases/latest", githubAPI, xrayRepo)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	tagName, ok := result["tag_name"].(string)
	if !ok {
		return "", fmt.Errorf("no tag_name in response")
	}

	return strings.TrimPrefix(tagName, "v"), nil
}

func downloadFile(url, dest string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()

	written, err := io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	if verbose {
		fmt.Printf("  Downloaded %d bytes\n", written)
	}

	return nil
}

func extractZip(zipPath, destDir string) error {
	if runtime.GOOS == "darwin" || runtime.GOOS == "linux" {
		return exec.Command("unzip", "-o", zipPath, "-d", destDir).Run()
	}
	return fmt.Errorf("zip extraction not supported on %s", runtime.GOOS)
}

func checksumFile(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}
