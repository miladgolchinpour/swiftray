package platform

import (
	"fmt"
	"os/exec"
)

type WindowsProxyManager struct{}

func NewWindowsProxyManager() *WindowsProxyManager {
	return &WindowsProxyManager{}
}

func (w *WindowsProxyManager) EnableProxy(httpPort, socksPort int) error {
	if httpPort > 0 {
		proxyAddr := fmt.Sprintf("127.0.0.1:%d", httpPort)
		cmd := exec.Command("reg", "add",
			`HKCU\Software\Microsoft\Windows\CurrentVersion\Internet Settings`,
			"/v", "ProxyEnable", "/t", "REG_DWORD", "/d", "1", "/f")
		if out, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("enable proxy: %s: %s", err, string(out))
		}

		cmd = exec.Command("reg", "add",
			`HKCU\Software\Microsoft\Windows\CurrentVersion\Internet Settings`,
			"/v", "ProxyServer", "/t", "REG_SZ", "/d", proxyAddr, "/f")
		if out, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("set proxy: %s: %s", err, string(out))
		}
	}

	return nil
}

func (w *WindowsProxyManager) DisableProxy() error {
	cmd := exec.Command("reg", "add",
		`HKCU\Software\Microsoft\Windows\CurrentVersion\Internet Settings`,
		"/v", "ProxyEnable", "/t", "REG_DWORD", "/d", "0", "/f")
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("disable proxy: %s: %s", err, string(out))
	}
	return nil
}
