package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	XrayVersion   string
	GeoIPURL      string
	GeoSiteURL    string
	XrayMirror    string
	OutputDir     string
	Platforms     []string
}

func DefaultConfig() Config {
	return Config{
		XrayVersion: "latest",
		GeoIPURL:    "https://github.com/XTLS/Xray-core/releases/latest/download/geoip.dat",
		GeoSiteURL:  "https://github.com/XTLS/Xray-core/releases/latest/download/geosite.dat",
		XrayMirror:  "https://github.com/XTLS/Xray-core/releases/latest/download",
		OutputDir:   "resources",
	}
}

func LoadConfig() Config {
	cfg := DefaultConfig()

	// Try .env.resmanager first, then .env
	for _, envFile := range []string{".env.resmanager", ".env"} {
		if _, err := os.Stat(envFile); err == nil {
			cfg.loadFromFile(envFile)
			break
		}
	}

	// Environment variables override file
	if v := os.Getenv("XRAY_VERSION"); v != "" {
		cfg.XrayVersion = v
	}
	if v := os.Getenv("XRAY_MIRROR"); v != "" {
		cfg.XrayMirror = v
	}
	if v := os.Getenv("GEO_IP_URL"); v != "" {
		cfg.GeoIPURL = v
	}
	if v := os.Getenv("GEO_SITE_URL"); v != "" {
		cfg.GeoSiteURL = v
	}
	if v := os.Getenv("RESOURCES_DIR"); v != "" {
		cfg.OutputDir = v
	}

	return cfg
}

func (c *Config) loadFromFile(path string) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		eqIdx := strings.Index(line, "=")
		if eqIdx < 0 {
			continue
		}

		key := strings.TrimSpace(line[:eqIdx])
		value := strings.TrimSpace(line[eqIdx+1:])

		// Remove surrounding quotes
		if len(value) >= 2 {
			if (value[0] == '"' && value[len(value)-1] == '"') ||
				(value[0] == '\'' && value[len(value)-1] == '\'') {
				value = value[1 : len(value)-1]
			}
		}

		switch key {
		case "XRAY_VERSION":
			c.XrayVersion = value
		case "XRAY_MIRROR":
			c.XrayMirror = value
		case "GEO_IP_URL":
			c.GeoIPURL = value
		case "GEO_SITE_URL":
			c.GeoSiteURL = value
		case "RESOURCES_DIR":
			c.OutputDir = value
		}
	}
}

func (c *Config) XrayZipURL(platform Platform) string {
	zipName := platform.Zip
	if c.XrayMirror != "" {
		return fmt.Sprintf("%s/%s", c.XrayMirror, zipName)
	}
	return fmt.Sprintf("https://github.com/XTLS/Xray-core/releases/latest/download/%s", zipName)
}

func (c *Config) ResourcePath(parts ...string) string {
	return filepath.Join(append([]string{c.OutputDir}, parts...)...)
}
