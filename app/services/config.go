package services

import (
	"strings"

	"SwiftRay/app/models"
	"SwiftRay/app/utils"
)

func GenerateConfig(node models.Node, settings models.AppSettings) map[string]interface{} {
	cfg := map[string]interface{}{
		"log":       generateLog(),
		"dns":       generateDNS(settings),
		"inbounds":  generateInbounds(settings),
		"outbounds": generateOutbounds(node, settings),
		"routing":   generateRouting(settings),
	}
	return cfg
}

func generateLog() map[string]interface{} {
	return map[string]interface{}{
		"loglevel": "warning",
		"access":   utils.AppDataDir() + "/access.log",
		"error":    utils.AppDataDir() + "/error.log",
	}
}

func generateDNS(settings models.AppSettings) map[string]interface{} {
	servers := []interface{}{}

	if settings.LocalDNS != "" {
		servers = append(servers, settings.LocalDNS)
	}

	if settings.RemoteDNS != "" {
		remoteServer := map[string]interface{}{
			"address": settings.RemoteDNS,
		}
		if settings.FakeIP {
			remoteServer["domains"] = []string{"geosite:geolocation-!cn"}
			remoteServer["expectIPs"] = []string{"geoip:!cn"}
		}
		servers = append(servers, remoteServer)
	}

	if settings.BootstrapDNS != "" && settings.BootstrapDNS != settings.LocalDNS {
		servers = append(servers, settings.BootstrapDNS)
	}

	dns := map[string]interface{}{
		"servers": servers,
	}

	if settings.FakeIP {
		dns["queryStrategy"] = "UseIP"
		dns["disableCache"] = false
		dns["disableFallback"] = false
	}

	return dns
}

func generateInbounds(settings models.AppSettings) []interface{} {
	inbounds := []interface{}{}

	listenAddr := "127.0.0.1"
	if settings.AllowLAN {
		listenAddr = "0.0.0.0"
	}

	sniffing := map[string]interface{}{
		"enabled": settings.EnableSniffing,
	}
	if settings.EnableSniffing {
		destOverride := []string{}
		if settings.SniffHTTP {
			destOverride = append(destOverride, "http")
		}
		if settings.SniffTLS {
			destOverride = append(destOverride, "tls")
		}
		if settings.SniffQUIC {
			destOverride = append(destOverride, "quic")
		}
		if settings.SniffFakeDNS {
			destOverride = append(destOverride, "fakedns")
		}
		sniffing["destOverride"] = destOverride
		sniffing["routeOnly"] = settings.RouteOnly
	}

	socksSettings := map[string]interface{}{
		"auth": "noauth",
		"udp":  settings.EnableUDP,
	}
	if settings.UseProxyAuth {
		socksSettings["auth"] = "password"
		socksSettings["accounts"] = []map[string]string{
			{
				"user": settings.ProxyUsername,
				"pass": settings.ProxyPassword,
			},
		}
	}

	inbounds = append(inbounds, map[string]interface{}{
		"tag":      "socks-in",
		"port":     settings.SOCKSPort,
		"listen":   listenAddr,
		"protocol": "socks",
		"settings": socksSettings,
		"sniffing": sniffing,
	})

	if settings.MixedPort {
		inbounds = append(inbounds, map[string]interface{}{
			"tag":      "mixed-in",
			"port":     settings.SOCKSPort,
			"listen":   listenAddr,
			"protocol": "http",
			"settings": map[string]interface{}{},
			"sniffing": sniffing,
		})
	} else {
		inbounds = append(inbounds, map[string]interface{}{
			"tag":      "http-in",
			"port":     settings.HTTPPort,
			"listen":   listenAddr,
			"protocol": "http",
			"settings": map[string]interface{}{},
		})
	}

	return inbounds
}

func generateOutbounds(node models.Node, settings models.AppSettings) []interface{} {
	outbounds := []interface{}{}

	outbound := generatePrimaryOutbound(node, settings)

	if settings.EnableFragment {
		outbound["streamSettings"] = addFragment(
			outbound["streamSettings"].(map[string]interface{}),
			settings,
		)
	}

	outbounds = append(outbounds, outbound)

	outbounds = append(outbounds, map[string]interface{}{
		"tag":      "direct",
		"protocol": "freedom",
		"settings": map[string]interface{}{},
	})

	outbounds = append(outbounds, map[string]interface{}{
		"tag":      "block",
		"protocol": "blackhole",
		"settings": map[string]interface{}{},
	})

	return outbounds
}

func generatePrimaryOutbound(node models.Node, settings models.AppSettings) map[string]interface{} {
	outbound := map[string]interface{}{
		"tag": "proxy",
	}

	switch node.ProtocolType {
	case models.ProtocolVMess:
		outbound["protocol"] = "vmess"
		outbound["settings"] = map[string]interface{}{
			"vnext": []interface{}{
				map[string]interface{}{
					"address": node.Address,
					"port":    node.Port,
					"users": []interface{}{
						map[string]interface{}{
							"id":       node.UUID,
							"alterId":  node.AlterID,
							"security": node.SecurityOrAuto(),
						},
					},
				},
			},
		}
	case models.ProtocolVLess:
		encryption := node.Encryption
		if encryption == "" {
			encryption = "none"
		}
		user := map[string]interface{}{
			"id":         node.UUID,
			"encryption": encryption,
		}
		if node.Flow != "" {
			user["flow"] = node.Flow
		}
		outbound["protocol"] = "vless"
		outbound["settings"] = map[string]interface{}{
			"vnext": []interface{}{
				map[string]interface{}{
					"address": node.Address,
					"port":    node.Port,
					"users":   []interface{}{user},
				},
			},
		}
	case models.ProtocolTrojan:
		outbound["protocol"] = "trojan"
		outbound["settings"] = map[string]interface{}{
			"servers": []interface{}{
				map[string]interface{}{
					"address":  node.Address,
					"port":     node.Port,
					"password": node.Password,
				},
			},
		}
	case models.ProtocolShadowsocks:
		outbound["protocol"] = "shadowsocks"
		outbound["settings"] = map[string]interface{}{
			"servers": []interface{}{
				map[string]interface{}{
					"address":  node.Address,
					"port":     node.Port,
					"method":   node.Cipher,
					"password": node.Password,
				},
			},
		}
	}

	outbound["streamSettings"] = generateStreamSettings(node, settings)

	return outbound
}

func generateStreamSettings(node models.Node, settings models.AppSettings) map[string]interface{} {
	stream := map[string]interface{}{
		"network": string(node.Transport),
	}

	if node.TLS {
		security := "tls"
		if node.RealityPublicKey != "" {
			security = "reality"
		}
		stream["security"] = security

		if security == "reality" {
			realitySettings := map[string]interface{}{}
			if node.SNI != "" {
				realitySettings["serverName"] = node.SNI
			}
			fp := node.Fingerprint
			if fp == "" {
				fp = settings.DefaultFingerprint
			}
			if fp != "" {
				realitySettings["fingerprint"] = fp
			}
			if node.RealityPublicKey != "" {
				realitySettings["publicKey"] = node.RealityPublicKey
			}
			if node.RealityShortID != "" {
				realitySettings["shortId"] = node.RealityShortID
			}
			if node.RealitySpiderX != "" {
				realitySettings["spiderX"] = node.RealitySpiderX
			}
			stream["realitySettings"] = realitySettings
		} else {
			tlsSettings := map[string]interface{}{}
			if node.SNI != "" {
				tlsSettings["serverName"] = node.SNI
			}
			fp := node.Fingerprint
			if fp == "" {
				fp = settings.DefaultFingerprint
			}
			if fp != "" {
				tlsSettings["fingerprint"] = fp
			}
			if node.ALPN != "" {
				tlsSettings["alpn"] = strings.Split(node.ALPN, ",")
			}
			stream["tlsSettings"] = tlsSettings
		}
	}

	switch node.Transport {
	case models.TransportWS:
		wsSettings := map[string]interface{}{}
		if node.Path != "" {
			wsSettings["path"] = node.Path
		}
		if node.Host != "" {
			wsSettings["headers"] = map[string]string{"Host": node.Host}
		}
		stream["wsSettings"] = wsSettings

	case models.TransportGRPC:
		grpcSettings := map[string]interface{}{}
		if node.ServiceName != "" {
			grpcSettings["serviceName"] = node.ServiceName
		}
		stream["grpcSettings"] = grpcSettings

	case models.TransportHTTP:
		httpSettings := map[string]interface{}{}
		if node.Path != "" {
			httpSettings["path"] = node.Path
		}
		if node.Host != "" {
			httpSettings["host"] = []string{node.Host}
		}
		stream["httpSettings"] = httpSettings

	case models.TransportHTTPUpgrade:
		httpSettings := map[string]interface{}{}
		if node.Path != "" {
			httpSettings["path"] = node.Path
		}
		if node.Host != "" {
			httpSettings["host"] = []string{node.Host}
		}
		stream["httpSettings"] = httpSettings

	case models.TransportSplitHTTP:
		httpSettings := map[string]interface{}{}
		if node.Path != "" {
			httpSettings["path"] = node.Path
		}
		if node.Host != "" {
			httpSettings["host"] = node.Host
		}
		stream["splithttpSettings"] = httpSettings

	case models.TransportXHTTP:
		httpSettings := map[string]interface{}{}
		if node.Path != "" {
			httpSettings["path"] = node.Path
		}
		if node.Host != "" {
			httpSettings["host"] = node.Host
		}
		stream["xhttpSettings"] = httpSettings
	}

	return stream
}

func addFragment(stream map[string]interface{}, settings models.AppSettings) map[string]interface{} {
	stream["fragment"] = map[string]interface{}{
		"pack":    "tlshello",
		"length":  settings.FragmentPackLength,
		"sleep":   settings.FragmentSleep,
	}
	return stream
}

func generateRouting(settings models.AppSettings) map[string]interface{} {
	routing := map[string]interface{}{}

	strategy := "AsIs"
	switch settings.DomainStrategy {
	case 1:
		strategy = "IPIfNonMatch"
	case 2:
		strategy = "IPOnDemand"
	}
	routing["domainStrategy"] = strategy

	if settings.RoutingMode == 0 {
		routing["rules"] = []interface{}{
			map[string]interface{}{
				"type":        "field",
				"outboundTag": "proxy",
				"ip":          []string{"0.0.0.0/0", "::/0"},
			},
		}
	} else {
		rules := []interface{}{
			map[string]interface{}{
				"type":        "field",
				"ip":          []string{"geoip:private"},
				"outboundTag": "direct",
			},
		}

		exclusions := splitAndTrimLines(settings.Exclusions)
		if len(exclusions) > 0 {
			rules = append(rules, map[string]interface{}{
				"type":        "field",
				"domain":      exclusions,
				"outboundTag": "direct",
			})
		}

		if settings.BypassIran {
			rules = append(rules, map[string]interface{}{
				"type":        "field",
				"domain":      []string{"geosite:geolocation-ir", "domain:.ir"},
				"outboundTag": "direct",
			})
			rules = append(rules, map[string]interface{}{
				"type":        "field",
				"ip":          []string{"geoip:ir"},
				"outboundTag": "direct",
			})
		}

		if settings.BypassRussia {
			rules = append(rules, map[string]interface{}{
				"type":        "field",
				"domain":      []string{"geosite:geolocation-ru", "domain:.ru"},
				"outboundTag": "direct",
			})
			rules = append(rules, map[string]interface{}{
				"type":        "field",
				"ip":          []string{"geoip:ru"},
				"outboundTag": "direct",
			})
		}

		if settings.BypassChina {
			rules = append(rules, map[string]interface{}{
				"type":        "field",
				"domain":      []string{"geosite:geolocation-cn", "domain:.cn"},
				"outboundTag": "direct",
			})
			rules = append(rules, map[string]interface{}{
				"type":        "field",
				"ip":          []string{"geoip:cn"},
				"outboundTag": "direct",
			})
		}

		rules = append(rules, map[string]interface{}{
			"type":        "field",
			"outboundTag": "proxy",
			"ip":          []string{"0.0.0.0/0", "::/0"},
		})

		routing["rules"] = rules
	}

	return routing
}

func splitAndTrimLines(s string) []string {
	lines := strings.Split(s, "\n")
	var result []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}
