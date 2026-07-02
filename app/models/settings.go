package models

type AppSettings struct {
	HTTPPort            int    `json:"httpPort"`
	SOCKSPort           int    `json:"socksPort"`
	MixedPort           bool   `json:"mixedPort"`
	EnableUDP           bool   `json:"enableUDP"`
	AllowLAN            bool   `json:"allowLAN"`
	RouteOnly           bool   `json:"routeOnly"`
	EnableSniffing      bool   `json:"enableSniffing"`
	SniffHTTP           bool   `json:"sniffHTTP"`
	SniffTLS            bool   `json:"sniffTLS"`
	SniffQUIC           bool   `json:"sniffQUIC"`
	SniffFakeDNS        bool   `json:"sniffFakeDNS"`
	UseProxyAuth        bool   `json:"useProxyAuth"`
	ProxyUsername        string `json:"proxyUsername"`
	ProxyPassword        string `json:"proxyPassword"`
	DefaultFingerprint   string `json:"defaultFingerprint"`
	EnableFragment       bool   `json:"enableFragment"`
	FragmentPackLength   string `json:"fragmentPackLength"`
	FragmentSleep        string `json:"fragmentSleep"`
	FragmentInterval     string `json:"fragmentInterval"`
	LocalDNS             string `json:"localDNS"`
	RemoteDNS            string `json:"remoteDNS"`
	BootstrapDNS         string `json:"bootstrapDNS"`
	ParallelQuery        bool   `json:"parallelQuery"`
	ServeStale           bool   `json:"serveStale"`
	UseSystemHosts       bool   `json:"useSystemHosts"`
	CustomDNSHosts       bool   `json:"customDNSHosts"`
	FakeIP               bool   `json:"fakeIP"`
	BlockSVCBHTTPS       bool   `json:"blockSVCBHTTPS"`
	ValidateRegionalDomain int  `json:"validateRegionalDomain"`
	EnableSystemProxy    bool   `json:"enableSystemProxy"`
	EnableMenuBar        bool   `json:"enableMenuBar"`
	RoutingMode          int    `json:"routingMode"`
	DomainStrategy       int    `json:"domainStrategy"`
	BypassIran           bool   `json:"bypassIran"`
	BypassRussia         bool   `json:"bypassRussia"`
	BypassChina          bool   `json:"bypassChina"`
	PingTestURL          string  `json:"pingTestURL"`
	CustomGeoSources     string  `json:"customGeoSources"`
	Exclusions           string  `json:"exclusions"`
	URLTestMode          string  `json:"urlTestMode"`
	URLTestTimeout       float64 `json:"urlTestTimeout"`
	URLTestConcurrency   int     `json:"urlTestConcurrency"`
}

func (s AppSettings) Port() int {
	if s.MixedPort {
		return s.SOCKSPort
	}
	return s.HTTPPort
}

func DefaultSettings() AppSettings {
	return AppSettings{
		HTTPPort:             2080,
		SOCKSPort:            2080,
		MixedPort:            true,
		EnableUDP:            true,
		AllowLAN:             true,
		RouteOnly:            false,
		EnableSniffing:       true,
		SniffHTTP:            true,
		SniffTLS:             true,
		SniffQUIC:            true,
		SniffFakeDNS:         false,
		UseProxyAuth:         false,
		ProxyUsername:         "",
		ProxyPassword:         "",
		DefaultFingerprint:    "chrome",
		EnableFragment:        false,
		FragmentPackLength:    "100-200",
		FragmentSleep:         "50-100",
		FragmentInterval:      "1-5",
		LocalDNS:              "8.8.8.8",
		RemoteDNS:             "https://cloudflare-dns.com/dns-query",
		BootstrapDNS:          "8.8.8.8",
		ParallelQuery:         true,
		ServeStale:            true,
		UseSystemHosts:        true,
		CustomDNSHosts:        false,
		FakeIP:                false,
		BlockSVCBHTTPS:        false,
		ValidateRegionalDomain: 0,
		EnableSystemProxy:     true,
		EnableMenuBar:         true,
		RoutingMode:           0,
		DomainStrategy:        1,
		BypassIran:            false,
		BypassRussia:          false,
		BypassChina:           false,
		PingTestURL:           "https://www.google.com/generate_204",
		CustomGeoSources:      "",
		Exclusions:            "localhost\n127.0.0.0/8\n::1",
		URLTestMode:           "tcp",
		URLTestTimeout:        3.0,
		URLTestConcurrency:    8,
	}
}
