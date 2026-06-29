package platform

type ProxyManager interface {
	EnableProxy(httpPort, socksPort int) error
	DisableProxy() error
}
