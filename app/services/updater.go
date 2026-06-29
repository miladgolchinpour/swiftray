package services

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"SwiftRay/app/utils"
)

type UpdateStatus struct {
	Type           string `json:"type"`
	CurrentVersion string `json:"currentVersion"`
	LatestVersion  string `json:"latestVersion"`
	HasUpdate      bool   `json:"hasUpdate"`
	ReleaseDate    string `json:"releaseDate,omitempty"`
	Error          string `json:"error,omitempty"`
}

type DownloadProgress struct {
	Type        string  `json:"type"`
	Stage       string  `json:"stage"`
	Progress    float64 `json:"progress"`
	Status      string  `json:"status"`
	URL         string  `json:"url,omitempty"`
	DestPath    string  `json:"destPath,omitempty"`
	Downloaded  int64   `json:"downloaded,omitempty"`
	Total       int64   `json:"total,omitempty"`
	Speed       float64 `json:"speed,omitempty"`
	ETA         float64 `json:"eta,omitempty"`
	Platform    string  `json:"platform,omitempty"`
	Arch        string  `json:"arch,omitempty"`
	Version     string  `json:"version,omitempty"`
	TargetVer   string  `json:"targetVersion,omitempty"`
	Error       string  `json:"error,omitempty"`
}

type UpdaterService struct {
	client       *http.Client
	mu           sync.Mutex
	downloading  bool
	cancelChan   chan struct{}
	onProgress   func(DownloadProgress)
	xrayVersion  string
}

func NewUpdaterService() *UpdaterService {
	return &UpdaterService{
		client: &http.Client{Timeout: 60 * time.Second},
	}
}

func (u *UpdaterService) OnProgress(fn func(DownloadProgress)) {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.onProgress = fn
}

func (u *UpdaterService) emitProgress(p DownloadProgress) {
	u.mu.Lock()
	cb := u.onProgress
	u.mu.Unlock()
	if cb != nil {
		cb(p)
	}
}

func (u *UpdaterService) IsDownloading() bool {
	u.mu.Lock()
	defer u.mu.Unlock()
	return u.downloading
}

func (u *UpdaterService) GetCurrentXrayVersion() string {
	versionFile := filepath.Join(utils.BundledXrayDir(), "VERSION")
	data, err := os.ReadFile(versionFile)
	if err != nil {
		return u.xrayVersion
	}
	return strings.TrimSpace(string(data))
}

func (u *UpdaterService) CheckXrayUpdate() UpdateStatus {
	status := UpdateStatus{Type: "runtime", CurrentVersion: u.GetCurrentXrayVersion()}

	latestVersion, releaseDate, err := u.getLatestXrayVersion()
	if err != nil {
		status.Error = err.Error()
		return status
	}

	status.LatestVersion = latestVersion
	status.ReleaseDate = releaseDate
	status.HasUpdate = compareVersions(status.CurrentVersion, latestVersion) < 0
	return status
}

func (u *UpdaterService) DownloadXray(onProgress func(float64)) error {
	u.mu.Lock()
	if u.downloading {
		u.mu.Unlock()
		return fmt.Errorf("download already in progress")
	}
	u.downloading = true
	u.cancelChan = make(chan struct{})
	u.mu.Unlock()

	defer func() {
		u.mu.Lock()
		u.downloading = false
		u.mu.Unlock()
	}()

	platform := u.currentPlatform()

	// Stage 1: Check latest release
	u.emitProgress(DownloadProgress{Type: "runtime", Stage: "checking", Progress: 0, Status: "Checking latest release...", Platform: platform.Name, Arch: platform.GOARCH})
	latestVersion, _, err := u.getLatestXrayVersion()
	if err != nil {
		return fmt.Errorf("GitHub API request failed: %w", err)
	}

	zipURL := fmt.Sprintf("https://github.com/XTLS/Xray-core/releases/download/v%s/%s", latestVersion, platform.ZipName)
	currentVersion := u.GetCurrentXrayVersion()

	// Stage 2: Find matching asset
	u.emitProgress(DownloadProgress{Type: "runtime", Stage: "finding", Progress: 0.05, Status: "Finding matching asset...", URL: zipURL, Platform: platform.Name, Arch: platform.GOARCH, Version: currentVersion, TargetVer: latestVersion})

	// Stage 3: Download ZIP
	u.emitProgress(DownloadProgress{Type: "runtime", Stage: "downloading", Progress: 0.1, Status: "Downloading Xray v" + latestVersion + "...", URL: zipURL, Platform: platform.Name, Arch: platform.GOARCH, Version: currentVersion, TargetVer: latestVersion})

	tmpDir := filepath.Join(os.TempDir(), "swiftray_update")
	os.MkdirAll(tmpDir, 0o755)
	tmpZip := filepath.Join(tmpDir, "xray.zip")

	startTime := time.Now()
	if err := u.downloadWithProgressDetailed(zipURL, tmpZip, func(downloaded, total int64, speed float64) {
		prog := 0.1
		if total > 0 {
			prog = 0.1 + 0.65*(float64(downloaded)/float64(total))
		}
		eta := float64(0)
		if speed > 0 && total > downloaded {
			eta = float64(total-downloaded) / speed
		}
		u.emitProgress(DownloadProgress{
			Type: "runtime", Stage: "downloading", Progress: prog,
			Status: fmt.Sprintf("Downloading... %s / %s", formatBytes(downloaded), formatBytes(total)),
			URL: zipURL, DestPath: tmpZip,
			Downloaded: downloaded, Total: total, Speed: speed, ETA: eta,
			Platform: platform.Name, Arch: platform.GOARCH, Version: currentVersion, TargetVer: latestVersion,
		})
	}); err != nil {
		os.Remove(tmpZip)
		return fmt.Errorf("download failed: %w", err)
	}
	downloadDuration := time.Since(startTime)

	// Stage 4: Verify archive
	u.emitProgress(DownloadProgress{Type: "runtime", Stage: "verifying", Progress: 0.75, Status: "Verifying archive...", Platform: platform.Name, Version: currentVersion, TargetVer: latestVersion})
	if info, err := os.Stat(tmpZip); err != nil || info.Size() < 1024 {
		os.Remove(tmpZip)
		return fmt.Errorf("downloaded file is invalid (%d bytes)", info.Size())
	}

	// Stage 5: Extract
	u.emitProgress(DownloadProgress{Type: "runtime", Stage: "extracting", Progress: 0.8, Status: "Extracting archive...", Platform: platform.Name, Version: currentVersion, TargetVer: latestVersion})
	extractDir := filepath.Join(tmpDir, "extract")
	os.MkdirAll(extractDir, 0o755)
	if err := u.extractZip(tmpZip, extractDir); err != nil {
		os.RemoveAll(tmpDir)
		return fmt.Errorf("ZIP extraction failed: %w", err)
	}

	// Stage 6: Replace runtime
	u.emitProgress(DownloadProgress{Type: "runtime", Stage: "replacing", Progress: 0.85, Status: "Replacing runtime files...", Platform: platform.Name, Version: currentVersion, TargetVer: latestVersion})

	xrayName := platform.XrayBinaryName()
	extracted := u.findExtractedBinary(extractDir, xrayName)
	if extracted == "" {
		os.RemoveAll(tmpDir)
		return fmt.Errorf("asset not found for current platform: xray binary missing in archive")
	}

	extractedGeoIP := u.findFile(extractDir, "geoip.dat")
	extractedGeoSite := u.findFile(extractDir, "geosite.dat")

	// Replace xray
	xrayDir := utils.BundledXrayDir()
	os.MkdirAll(xrayDir, 0o755)
	finalXray := filepath.Join(xrayDir, xrayName)
	backupXray := finalXray + ".bak"
	os.Remove(backupXray)
	os.Rename(finalXray, backupXray)
	if err := os.Rename(extracted, finalXray); err != nil {
		os.Rename(backupXray, finalXray)
		os.RemoveAll(tmpDir)
		return fmt.Errorf("failed to replace xray binary: %w", err)
	}

	// Stage 7: Set permissions
	u.emitProgress(DownloadProgress{Type: "runtime", Stage: "permissions", Progress: 0.9, Status: "Setting executable permissions...", Platform: platform.Name, Version: currentVersion, TargetVer: latestVersion})
	if runtime.GOOS != "windows" {
		if err := os.Chmod(finalXray, 0o755); err != nil {
			os.RemoveAll(tmpDir)
			return fmt.Errorf("failed to set executable permissions: %w", err)
		}
	}

	// Replace geo files
	geoDir := utils.BundledGeoDir()
	os.MkdirAll(geoDir, 0o755)

	for _, gf := range []struct {
		extracted string
		name      string
	}{
		{extractedGeoIP, "geoip.dat"},
		{extractedGeoSite, "geosite.dat"},
	} {
		if gf.extracted == "" {
			continue
		}
		finalGeo := filepath.Join(geoDir, gf.name)
		backupGeo := finalGeo + ".bak"
		os.Remove(backupGeo)
		os.Rename(finalGeo, backupGeo)
		if err := os.Rename(gf.extracted, finalGeo); err != nil {
			os.Rename(backupGeo, finalGeo)
			os.RemoveAll(tmpDir)
			return fmt.Errorf("failed to replace %s: %w", gf.name, err)
		}
		os.Remove(backupGeo)
	}

	// Stage 8: Validate installation
	u.emitProgress(DownloadProgress{Type: "runtime", Stage: "validating", Progress: 0.95, Status: "Validating installation...", Platform: platform.Name, Version: currentVersion, TargetVer: latestVersion})

	if _, err := os.Stat(finalXray); err != nil {
		os.RemoveAll(tmpDir)
		return fmt.Errorf("validation failed: xray binary not found after install")
	}
	if _, err := os.Stat(filepath.Join(geoDir, "geoip.dat")); err != nil {
		os.RemoveAll(tmpDir)
		return fmt.Errorf("validation failed: geoip.dat not found after install")
	}
	if _, err := os.Stat(filepath.Join(geoDir, "geosite.dat")); err != nil {
		os.RemoveAll(tmpDir)
		return fmt.Errorf("validation failed: geosite.dat not found after install")
	}

	// Write version files
	os.WriteFile(filepath.Join(xrayDir, "VERSION"), []byte(latestVersion), 0o644)
	os.WriteFile(filepath.Join(geoDir, "VERSION"), []byte(latestVersion), 0o644)

	// Cleanup
	os.Remove(backupXray)
	os.RemoveAll(tmpDir)

	u.xrayVersion = latestVersion
	u.emitProgress(DownloadProgress{
		Type: "runtime", Stage: "completed", Progress: 1.0,
		Status: fmt.Sprintf("Xray updated to v%s (%.1fs download)", latestVersion, downloadDuration.Seconds()),
		Platform: platform.Name, Version: currentVersion, TargetVer: latestVersion,
	})
	log.Printf("[Updater] Xray updated to v%s in %v", latestVersion, downloadDuration)
	return nil
}

func (u *UpdaterService) findFile(dir, name string) string {
	path := filepath.Join(dir, name)
	if _, err := os.Stat(path); err == nil {
		return path
	}
	entries, _ := os.ReadDir(dir)
	for _, e := range entries {
		if e.IsDir() {
			p := filepath.Join(dir, e.Name(), name)
			if _, err := os.Stat(p); err == nil {
				return p
			}
		}
	}
	return ""
}

func (u *UpdaterService) Cancel() {
	u.mu.Lock()
	ch := u.cancelChan
	u.mu.Unlock()
	if ch != nil {
		close(ch)
	}
}

func (u *UpdaterService) getLatestXrayVersion() (string, string, error) {
	url := "https://api.github.com/repos/XTLS/Xray-core/releases/latest"
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := u.client.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", "", err
	}

	tagName, _ := result["tag_name"].(string)
	publishedAt, _ := result["published_at"].(string)

	return strings.TrimPrefix(tagName, "v"), publishedAt, nil
}

type platformInfo struct {
	Name    string
	GOOS    string
	GOARCH  string
	ZipName string
}

func (p platformInfo) XrayBinaryName() string {
	if p.GOOS == "windows" {
		return "xray.exe"
	}
	return "xray"
}

func (u *UpdaterService) currentPlatform() platformInfo {
	switch {
	case runtime.GOOS == "darwin" && runtime.GOARCH == "arm64":
		return platformInfo{"darwin-arm64-v8a", "darwin", "arm64", "Xray-macos-arm64-v8a.zip"}
	case runtime.GOOS == "darwin" && runtime.GOARCH == "amd64":
		return platformInfo{"darwin-amd64", "darwin", "amd64", "Xray-macos-64.zip"}
	case runtime.GOOS == "linux" && runtime.GOARCH == "arm64":
		return platformInfo{"linux-arm64", "linux", "arm64", "Xray-linux-arm64-v8a.zip"}
	case runtime.GOOS == "linux" && runtime.GOARCH == "amd64":
		return platformInfo{"linux-amd64", "linux", "amd64", "Xray-linux-64.zip"}
	case runtime.GOOS == "windows" && runtime.GOARCH == "arm64":
		return platformInfo{"windows-arm64", "windows", "arm64", "Xray-windows-arm64-v8a.zip"}
	case runtime.GOOS == "windows":
		return platformInfo{"windows-amd64", "windows", "amd64", "Xray-windows-64.zip"}
	default:
		return platformInfo{"linux-amd64", "linux", "amd64", "Xray-linux-64.zip"}
	}
}

func (u *UpdaterService) downloadWithProgressDetailed(url, dest string, onProgress func(downloaded, total int64, speed float64)) error {
	resp, err := u.client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("HTTP %d from %s", resp.StatusCode, url)
	}

	total := resp.ContentLength
	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()

	buf := make([]byte, 64*1024)
	var written int64
	startTime := time.Now()
	lastReport := startTime

	for {
		select {
		case <-u.cancelChan:
			return fmt.Errorf("download cancelled")
		default:
		}

		n, readErr := resp.Body.Read(buf)
		if n > 0 {
			if _, err := out.Write(buf[:n]); err != nil {
				return err
			}
			written += int64(n)

			now := time.Now()
			if now.Sub(lastReport) >= 200*time.Millisecond || readErr == io.EOF {
				elapsed := now.Sub(startTime).Seconds()
				speed := float64(written) / elapsed
				if onProgress != nil {
					onProgress(written, total, speed)
				}
				lastReport = now
			}
		}
		if readErr != nil {
			if readErr == io.EOF {
				break
			}
			return readErr
		}
	}
	return nil
}

func formatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}

func (u *UpdaterService) extractZip(zipPath, destDir string) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		fpath := filepath.Join(destDir, f.Name)
		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, 0o755)
			continue
		}
		if err := os.MkdirAll(filepath.Dir(fpath), 0o755); err != nil {
			return err
		}
		outFile, err := os.Create(fpath)
		if err != nil {
			return err
		}
		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			return err
		}
		_, err = io.Copy(outFile, rc)
		rc.Close()
		outFile.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

func (u *UpdaterService) findExtractedBinary(dir, name string) string {
	if _, err := os.Stat(filepath.Join(dir, name)); err == nil {
		return filepath.Join(dir, name)
	}
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

// compareVersions compares two semantic version strings.
// Returns -1 if a < b, 0 if a == b, 1 if a > b.
func compareVersions(a, b string) int {
	aParts := parseVersion(a)
	bParts := parseVersion(b)

	for i := 0; i < 3; i++ {
		if aParts[i] < bParts[i] {
			return -1
		}
		if aParts[i] > bParts[i] {
			return 1
		}
	}
	return 0
}

func parseVersion(v string) [3]int {
	v = strings.TrimPrefix(v, "v")
	var parts [3]int
	for i := 0; i < 3; i++ {
		idx := strings.Index(v, ".")
		if idx < 0 {
			parts[i], _ = strconv.Atoi(v)
			break
		}
		parts[i], _ = strconv.Atoi(v[:idx])
		v = v[idx+1:]
	}
	return parts
}
