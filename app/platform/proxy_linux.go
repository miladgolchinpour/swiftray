package platform

import (
	"fmt"
	"os/exec"
)

type LinuxProxyManager struct{}

func NewLinuxProxyManager() *LinuxProxyManager {
	return &LinuxProxyManager{}
}

func (l *LinuxProxyManager) EnableProxy(httpPort, socksPort int) error {
	if err := l.setGSettings("org.gnome.system.proxy", "mode", "'manual'"); err != nil {
		return err
	}

	if httpPort > 0 {
		if err := l.setGSettings("org.gnome.system.proxy.http", "host", "'127.0.0.1'"); err != nil {
			return err
		}
		if err := l.setGSettingsInt("org.gnome.system.proxy.http", "port", httpPort); err != nil {
			return err
		}
	}

	if socksPort > 0 {
		if err := l.setGSettings("org.gnome.system.proxy.socks", "host", "'127.0.0.1'"); err != nil {
			return err
		}
		if err := l.setGSettingsInt("org.gnome.system.proxy.socks", "port", socksPort); err != nil {
			return err
		}
	}

	return nil
}

func (l *LinuxProxyManager) DisableProxy() error {
	return l.setGSettings("org.gnome.system.proxy", "mode", "'none'")
}

func (l *LinuxProxyManager) setGSettings(schema, key, value string) error {
	cmd := exec.Command("gsettings", "set", schema, key, value)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("gsettings error: %s: %s", err, string(out))
	}
	return nil
}

func (l *LinuxProxyManager) setGSettingsInt(schema, key string, value int) error {
	cmd := exec.Command("gsettings", "set", schema, key, fmt.Sprintf("%d", value))
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("gsettings int error: %s: %s", err, string(out))
	}
	return nil
}
