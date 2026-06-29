//go:build linux

package platform

func NewProxyManager() ProxyManager {
	return NewLinuxProxyManager()
}
