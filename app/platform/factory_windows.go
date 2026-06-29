//go:build windows

package platform

func NewProxyManager() ProxyManager {
	return NewWindowsProxyManager()
}
