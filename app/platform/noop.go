package platform

type noopProxyManager struct{}

func (n *noopProxyManager) EnableProxy(httpPort, socksPort int) error { return nil }
func (n *noopProxyManager) DisableProxy() error                      { return nil }
