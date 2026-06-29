package platform

import (
	"fmt"
	"os/exec"
	"strings"
)

type MacProxyManager struct{}

func NewMacProxyManager() *MacProxyManager {
	return &MacProxyManager{}
}

func (m *MacProxyManager) EnableProxy(httpPort, socksPort int) error {
	service, err := m.activeNetworkService()
	if err != nil {
		return err
	}

	if err := m.run("-setwebproxy", service, "127.0.0.1", itoa(httpPort)); err != nil {
		return err
	}
	if err := m.run("-setsecurewebproxy", service, "127.0.0.1", itoa(httpPort)); err != nil {
		return err
	}
	if err := m.run("-setsocksfirewallproxy", service, "127.0.0.1", itoa(socksPort)); err != nil {
		return err
	}
	if err := m.run("-setwebproxystate", service, "on"); err != nil {
		return err
	}
	if err := m.run("-setsecurewebproxystate", service, "on"); err != nil {
		return err
	}
	if err := m.run("-setsocksfirewallproxystate", service, "on"); err != nil {
		return err
	}

	return nil
}

func (m *MacProxyManager) DisableProxy() error {
	service, err := m.activeNetworkService()
	if err != nil {
		return err
	}

	var lastErr error
	if err := m.run("-setwebproxystate", service, "off"); err != nil {
		lastErr = err
	}
	if err := m.run("-setsecurewebproxystate", service, "off"); err != nil {
		lastErr = err
	}
	if err := m.run("-setsocksfirewallproxystate", service, "off"); err != nil {
		lastErr = err
	}

	return lastErr
}

func (m *MacProxyManager) run(args ...string) error {
	cmd := exec.Command("/usr/sbin/networksetup", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("networksetup error: %s: %s", err, string(out))
	}
	return nil
}

func (m *MacProxyManager) activeNetworkService() (string, error) {
	out, err := exec.Command("/usr/sbin/networksetup", "-listallnetworkservices").Output()
	if err != nil {
		return "", err
	}

	lines := strings.Split(string(out), "\n")
	var services []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.Contains(line, "denotes") {
			continue
		}
		services = append(services, line)
	}

	for _, s := range services {
		if s == "Wi-Fi" {
			return "Wi-Fi", nil
		}
	}
	for _, s := range services {
		if s == "Ethernet" {
			return "Ethernet", nil
		}
	}
	if len(services) > 0 {
		return services[0], nil
	}
	return "Wi-Fi", nil
}

func itoa(i int) string {
	return fmt.Sprintf("%d", i)
}
