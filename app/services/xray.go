package services

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"SwiftRay/app/models"
	"SwiftRay/app/utils"
)

type XrayService struct {
	storage     *StorageService
	mu          sync.Mutex
	process     *exec.Cmd
	logCallback func(level, message string)
}

func NewXrayService(storage *StorageService) *XrayService {
	return &XrayService{storage: storage}
}

func (x *XrayService) SetLogCallback(fn func(level, message string)) {
	x.mu.Lock()
	defer x.mu.Unlock()
	x.logCallback = fn
}

func (x *XrayService) EnsureDirectory() error {
	return os.MkdirAll(utils.AppDataDir(), 0o755)
}

func (x *XrayService) ValidateResources() error {
	return utils.ValidateBundledResources()
}

func (x *XrayService) GetResourceStatus() utils.BundledResourceStatus {
	return utils.GetBundledResourceStatus()
}

func (x *XrayService) IsInstalled() bool {
	_, err := os.Stat(utils.BundledXrayPath())
	return err == nil
}

func (x *XrayService) GetInstalledVersion() string {
	if !x.IsInstalled() {
		return ""
	}
	out, err := exec.Command(utils.BundledXrayPath(), "version").Output()
	if err != nil {
		return ""
	}
	output := string(out)
	if idx := strings.Index(output, "Xray "); idx >= 0 {
		after := output[idx+len("Xray "):]
		end := strings.IndexAny(after, " \n")
		if end >= 0 {
			return after[:end]
		}
		return after
	}
	return ""
}

func (x *XrayService) CheckForUpdate() (bool, error) {
	resp, err := httpGet("https://api.github.com/repos/XTLS/Xray-core/releases/latest")
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false, err
	}

	tagName, ok := result["tag_name"].(string)
	if !ok {
		return false, fmt.Errorf("no tag_name in response")
	}
	latestVersion := strings.TrimPrefix(tagName, "v")

	currentVersion := x.GetInstalledVersion()
	return currentVersion != latestVersion, nil
}

func (x *XrayService) WriteConfig(config map[string]interface{}) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(utils.ConfigPath(), data, 0o644)
}

func (x *XrayService) Start(onLogLine func(level, message string)) error {
	x.Stop()

	if !x.IsInstalled() {
		return fmt.Errorf("xray binary not found")
	}

	bin := utils.BundledXrayPath()
	cmd := exec.Command(bin, "run", "-c", utils.ConfigPath())

	env := os.Environ()
	env = append(env, "XRAY_LOCATION_ASSET="+utils.BundledGeoDir())
	cmd.Env = env

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("stdout pipe: %w", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start process: %w", err)
	}

	x.mu.Lock()
	x.process = cmd
	x.mu.Unlock()

	go x.readOutput(stdout, false, onLogLine)
	go x.readOutput(stderr, true, onLogLine)

	go func() {
		cmd.Wait()
		x.mu.Lock()
		if x.process == cmd {
			x.process = nil
		}
		x.mu.Unlock()
	}()

	return nil
}

func (x *XrayService) Stop() {
	x.mu.Lock()
	proc := x.process
	x.mu.Unlock()

	if proc != nil && proc.Process != nil {
		if proc.ProcessState == nil || !proc.ProcessState.Exited() {
			proc.Process.Signal(os.Interrupt)
			done := make(chan struct{})
			go func() {
				proc.Wait()
				close(done)
			}()
			select {
			case <-done:
			case <-time.After(5 * time.Second):
				proc.Process.Kill()
			}
		}
	}

	x.mu.Lock()
	x.process = nil
	x.mu.Unlock()
}

func (x *XrayService) IsRunning() bool {
	x.mu.Lock()
	proc := x.process
	x.mu.Unlock()

	if proc == nil || proc.Process == nil {
		return false
	}
	if proc.ProcessState != nil {
		return !proc.ProcessState.Exited()
	}
	return true
}

func (x *XrayService) readOutput(r io.Reader, isError bool, onLogLine func(level, message string)) {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		var level string
		if isError {
			if strings.Contains(strings.ToLower(line), "warning") {
				level = "warning"
			} else {
				level = "error"
			}
		} else {
			lower := strings.ToLower(line)
			switch {
			case strings.Contains(lower, "warning"):
				level = "warning"
			case strings.Contains(lower, "debug"):
				level = "debug"
			default:
				level = "info"
			}
		}

		if onLogLine != nil {
			onLogLine(level, line)
		}

		// Also forward to the global log callback
		x.mu.Lock()
		cb := x.logCallback
		x.mu.Unlock()
		if cb != nil {
			cb(level, line)
		}
	}
}

func (x *XrayService) TestLatencyConnect(host string, port int, timeout time.Duration) *int {
	addr := net.JoinHostPort(host, fmt.Sprintf("%d", port))
	start := time.Now()
	conn, err := net.DialTimeout("tcp", addr, timeout)
	if err != nil {
		return nil
	}
	latency := int(time.Since(start).Milliseconds())
	conn.Close()
	return &latency
}

func (x *XrayService) TestLatencyHTTP(host string, port int, timeout time.Duration) *int {
	addr := fmt.Sprintf("http://%s:%d", host, port)
	client := &http.Client{
		Timeout: timeout,
		Transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout:   timeout,
				DualStack: true,
			}).DialContext,
			TLSHandshakeTimeout:   timeout,
			ResponseHeaderTimeout:  timeout,
			DisableKeepAlives:      true,
			DisableCompression:     true,
			MaxIdleConns:           1,
			MaxIdleConnsPerHost:    1,
			IdleConnTimeout:        timeout,
		},
	}
	start := time.Now()
	resp, err := client.Get(addr)
	if err != nil {
		return nil
	}
	resp.Body.Close()
	latency := int(time.Since(start).Milliseconds())
	return &latency
}

func (x *XrayService) URLTest(nodes []models.Node, timeout float64, concurrency int, mode string, onNodeComplete func(nodeID string, delay *int)) {
	var wg sync.WaitGroup
	sem := make(chan struct{}, concurrency)

	for _, node := range nodes {
		wg.Add(1)
		sem <- struct{}{}
		go func(n models.Node) {
			defer wg.Done()
			defer func() { <-sem }()

			var dur *int
			if mode == "http" {
				dur = x.TestLatencyHTTP(n.Address, n.Port, time.Duration(timeout*float64(time.Second)))
			} else {
				dur = x.TestLatencyConnect(n.Address, n.Port, time.Duration(timeout*float64(time.Second)))
			}
			onNodeComplete(n.ID, dur)
		}(node)
	}

	wg.Wait()
}

func httpGet(url string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	return http.DefaultClient.Do(req)
}
