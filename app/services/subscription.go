package services

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"SwiftRay/app/models"
)

type SubscriptionService struct {
	client *http.Client
}

func NewSubscriptionService() *SubscriptionService {
	return &SubscriptionService{
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

func (s *SubscriptionService) Fetch(subURL string) ([]models.Node, error) {
	resp, err := s.client.Get(subURL)
	if err != nil {
		return nil, fmt.Errorf("fetch failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}

	return s.ParseContent(string(body))
}

func (s *SubscriptionService) ParseContent(content string) ([]models.Node, error) {
	content = strings.TrimSpace(content)

	if strings.HasPrefix(content, "{") {
		return s.parseJSON(content)
	}

	lines := strings.Split(content, "\n")
	var nodes []models.Node

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if isConfigURL(line) {
			node, err := s.parseConfigURL(line)
			if err == nil {
				nodes = append(nodes, node)
			}
			continue
		}

		if s.isBase64Content(line) {
			decoded, err := decodeBase64Safe(line)
			if err != nil {
				continue
			}
			subLines := strings.Split(decoded, "\n")
			for _, subLine := range subLines {
				subLine = strings.TrimSpace(subLine)
				if isConfigURL(subLine) {
					node, err := s.parseConfigURL(subLine)
					if err == nil {
						nodes = append(nodes, node)
					}
				}
			}
		}
	}

	return nodes, nil
}

func (s *SubscriptionService) FetchSync(subURL string, linkOverride string) []models.Node {
	if linkOverride != "" {
		nodes, _ := s.ParseContent(linkOverride)
		return nodes
	}
	if subURL == "" {
		return nil
	}
	resp, err := s.client.Get(subURL)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil
	}
	nodes, _ := s.ParseContent(string(body))
	return nodes
}

func (s *SubscriptionService) parseJSON(content string) ([]models.Node, error) {
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(content), &result); err != nil {
		return nil, err
	}

	var outbounds []interface{}
	if list, ok := result["list"].([]interface{}); ok {
		outbounds = list
	} else if obs, ok := result["outbounds"].([]interface{}); ok {
		outbounds = obs
	}

	var nodes []models.Node
	for _, ob := range outbounds {
		obMap, ok := ob.(map[string]interface{})
		if !ok {
			continue
		}
		node, err := s.parseJSONOutbound(obMap)
		if err == nil {
			nodes = append(nodes, node)
		}
	}

	return nodes, nil
}

func (s *SubscriptionService) parseJSONOutbound(ob map[string]interface{}) (models.Node, error) {
	node := models.NewNode()

	if tag, ok := ob["tag"].(string); ok {
		node.Name = tag
	}

	protocol, _ := ob["protocol"].(string)
	switch protocol {
	case "vmess":
		node.ProtocolType = models.ProtocolVMess
	case "vless":
		node.ProtocolType = models.ProtocolVLess
	case "trojan":
		node.ProtocolType = models.ProtocolTrojan
	case "shadowsocks":
		node.ProtocolType = models.ProtocolShadowsocks
	}

	settings, _ := ob["settings"].(map[string]interface{})

	switch protocol {
	case "vmess", "vless":
		if vnextList, ok := settings["vnext"].([]interface{}); ok && len(vnextList) > 0 {
			vnext, _ := vnextList[0].(map[string]interface{})
			node.Address, _ = vnext["address"].(string)
			if port, ok := vnext["port"].(float64); ok {
				node.Port = int(port)
			}
			if users, ok := vnext["users"].([]interface{}); ok && len(users) > 0 {
				user, _ := users[0].(map[string]interface{})
				node.UUID, _ = user["id"].(string)
				node.Security, _ = user["security"].(string)
				if aid, ok := user["alterId"].(float64); ok {
					node.AlterID = int(aid)
				}
				node.Flow, _ = user["flow"].(string)
				node.Encryption, _ = user["encryption"].(string)
			}
		}
	case "trojan", "shadowsocks":
		if servers, ok := settings["servers"].([]interface{}); ok && len(servers) > 0 {
			server, _ := servers[0].(map[string]interface{})
			node.Address, _ = server["address"].(string)
			if port, ok := server["port"].(float64); ok {
				node.Port = int(port)
			}
			node.Password, _ = server["password"].(string)
			node.Cipher, _ = server["method"].(string)
		}
	}

	if stream, ok := ob["streamSettings"].(map[string]interface{}); ok {
		node.Transport = models.NormalizeTransport(models.NodeTransport(fmt.Sprintf("%v", stream["network"])))
		if sec, ok := stream["security"].(string); ok && sec == "tls" {
			node.TLS = true
			if tlsSettings, ok := stream["tlsSettings"].(map[string]interface{}); ok {
				node.SNI, _ = tlsSettings["serverName"].(string)
				node.Fingerprint, _ = tlsSettings["fingerprint"].(string)
			}
		}
		if sec, ok := stream["security"].(string); ok && sec == "reality" {
			if realitySettings, ok := stream["realitySettings"].(map[string]interface{}); ok {
				node.SNI, _ = realitySettings["serverName"].(string)
				node.RealityPublicKey, _ = realitySettings["publicKey"].(string)
				node.RealityShortID, _ = realitySettings["shortId"].(string)
				node.RealitySpiderX, _ = realitySettings["spiderX"].(string)
			}
		}
	}

	return node, nil
}

func (s *SubscriptionService) parseConfigURL(raw string) (models.Node, error) {
	prefix := strings.SplitN(raw, "://", 2)
	if len(prefix) != 2 {
		return models.Node{}, fmt.Errorf("invalid url")
	}

	scheme := strings.ToLower(prefix[0])
	rest := prefix[1]

	switch scheme {
	case "vmess":
		return s.parseVMess(rest)
	case "vless":
		return s.parseVLess(rest)
	case "trojan":
		return s.parseTrojan(rest)
	case "ss":
		return s.parseShadowsocks(rest)
	default:
		return models.Node{}, fmt.Errorf("unsupported protocol: %s", scheme)
	}
}

func (s *SubscriptionService) parseVMess(encoded string) (models.Node, error) {
	decoded, err := decodeBase64Safe(encoded)
	if err != nil {
		return models.Node{}, fmt.Errorf("decode vmess: %w", err)
	}

	var data map[string]interface{}
	if err := json.Unmarshal([]byte(decoded), &data); err != nil {
		return models.Node{}, fmt.Errorf("parse vmess json: %w", err)
	}

	node := models.NewNode()
	node.ProtocolType = models.ProtocolVMess
	node.Name = getString(data, "ps")
	node.Address = getString(data, "add")
	node.Port = getInt(data, "port")
	node.UUID = getString(data, "id")
	node.AlterID = getInt(data, "aid")
	node.Security = getString(data, "scy")
	if node.Security == "" {
		node.Security = "auto"
	}
	node.Transport = models.NodeTransport(getString(data, "net"))
	node.Host = getString(data, "host")
	node.Path = getString(data, "path")
	node.SNI = getString(data, "sni")
	node.ALPN = getString(data, "alpn")
	node.Fingerprint = getString(data, "fp")

	tlsStr := getString(data, "tls")
	if tlsStr == "tls" {
		node.TLS = true
	}

	if node.Transport == "" {
		node.Transport = models.TransportTCP
	}
	if node.Port == 0 {
		node.Port = 443
	}

	node.RawLink = "vmess://" + encoded
	return node, nil
}

func (s *SubscriptionService) parseVLess(raw string) (models.Node, error) {
	node := models.NewNode()
	node.ProtocolType = models.ProtocolVLess

	atIdx := strings.Index(raw, "@")
	hashIdx := strings.Index(raw, "#")
	queryIdx := strings.Index(raw, "?")

	if atIdx < 0 {
		return node, fmt.Errorf("invalid vless url")
	}

	node.UUID = raw[:atIdx]

	hostPort := raw[atIdx+1:]
	if queryIdx > 0 && queryIdx < hashIdx {
		hostPort = raw[atIdx+1 : queryIdx]
	} else if hashIdx > 0 {
		hostPort = raw[atIdx+1 : hashIdx]
	}

	parts := strings.Split(hostPort, ":")
	if len(parts) != 2 {
		return node, fmt.Errorf("invalid host:port")
	}
	node.Address = parts[0]
	node.Port, _ = strconv.Atoi(parts[1])

	if hashIdx > 0 {
		node.Name = unescapeName(raw[hashIdx+1:])
	}

	if queryIdx > 0 {
		params := raw[queryIdx:]
		if hashIdx > queryIdx {
			params = raw[queryIdx:hashIdx]
		}
		q, _ := url.ParseQuery(strings.TrimPrefix(params, "?"))
		if t := q.Get("type"); t != "" {
			node.Transport = models.NormalizeTransport(models.NodeTransport(t))
		}
		security := q.Get("security")
		node.TLS = security == "tls" || security == "reality"
		if !node.TLS && q.Get("sni") != "" {
			node.TLS = true
		}
		node.SNI = q.Get("sni")
		node.Fingerprint = q.Get("fp")
		node.ALPN = q.Get("alpn")
		node.Flow = q.Get("flow")
		node.Encryption = q.Get("encryption")
		node.Host = q.Get("host")
		node.Path = q.Get("path")
		node.ServiceName = q.Get("serviceName")
		node.RealityPublicKey = q.Get("pbk")
		node.RealityShortID = q.Get("sid")
		node.RealitySpiderX = q.Get("spx")
	}

	node.RawLink = "vless://" + raw
	return node, nil
}

func (s *SubscriptionService) parseTrojan(raw string) (models.Node, error) {
	node := models.NewNode()
	node.ProtocolType = models.ProtocolTrojan

	atIdx := strings.Index(raw, "@")
	hashIdx := strings.Index(raw, "#")
	queryIdx := strings.Index(raw, "?")

	if atIdx < 0 {
		return node, fmt.Errorf("invalid trojan url")
	}

	node.Password = raw[:atIdx]

	hostPort := raw[atIdx+1:]
	if queryIdx > 0 && queryIdx < hashIdx {
		hostPort = raw[atIdx+1 : queryIdx]
	} else if hashIdx > 0 {
		hostPort = raw[atIdx+1 : hashIdx]
	}

	parts := strings.Split(hostPort, ":")
	if len(parts) != 2 {
		return node, fmt.Errorf("invalid host:port")
	}
	node.Address = parts[0]
	node.Port, _ = strconv.Atoi(parts[1])

	if hashIdx > 0 {
		node.Name = unescapeName(raw[hashIdx+1:])
	}

	if queryIdx > 0 {
		params := raw[queryIdx:]
		if hashIdx > queryIdx {
			params = raw[queryIdx:hashIdx]
		}
		q, _ := url.ParseQuery(strings.TrimPrefix(params, "?"))
		if t := q.Get("type"); t != "" {
			node.Transport = models.NormalizeTransport(models.NodeTransport(t))
		}
		security := q.Get("security")
		if security == "" {
			security = "tls"
		}
		node.TLS = security == "tls" || security == "reality"
		node.SNI = q.Get("sni")
		node.Fingerprint = q.Get("fp")
		node.ALPN = q.Get("alpn")
		node.Host = q.Get("host")
		node.Path = q.Get("path")
		node.ServiceName = q.Get("serviceName")
	} else {
		node.TLS = true
	}

	node.RawLink = "trojan://" + raw
	return node, nil
}

func (s *SubscriptionService) parseShadowsocks(raw string) (models.Node, error) {
	node := models.NewNode()
	node.ProtocolType = models.ProtocolShadowsocks

	atIdx := strings.Index(raw, "@")
	hashIdx := strings.Index(raw, "#")

	if atIdx > 0 {
		cipherPass := raw[:atIdx]
		decoded, err := decodeBase64Safe(cipherPass)
		if err == nil {
			parts := strings.SplitN(decoded, ":", 2)
			if len(parts) == 2 {
				node.Cipher = parts[0]
				node.Password = parts[1]
			}
		}

		hostPort := raw[atIdx+1:]
		if hashIdx > 0 {
			hostPort = raw[atIdx+1 : hashIdx]
		}
		parts := strings.Split(hostPort, ":")
		if len(parts) == 2 {
			node.Address = parts[0]
			node.Port, _ = strconv.Atoi(parts[1])
		}
	} else {
		questionIdx := strings.Index(raw, "?")
		mainPart := raw
		if questionIdx > 0 {
			mainPart = raw[:questionIdx]
		}
		if hashIdx > 0 {
			mainPart = raw[:hashIdx]
		}

		decoded, err := decodeBase64Safe(mainPart)
		if err == nil {
			atIdx2 := strings.LastIndex(decoded, "@")
			if atIdx2 > 0 {
				cipherPass := decoded[:atIdx2]
				parts := strings.SplitN(cipherPass, ":", 2)
				if len(parts) == 2 {
					node.Cipher = parts[0]
					node.Password = parts[1]
				}
				hostPort := decoded[atIdx2+1:]
				hpParts := strings.Split(hostPort, ":")
				if len(hpParts) == 2 {
					node.Address = hpParts[0]
					node.Port, _ = strconv.Atoi(hpParts[1])
				}
			}
		}
	}

	if hashIdx > 0 {
		node.Name = unescapeName(raw[hashIdx+1:])
	}

	if node.Port == 0 {
		node.Port = 443
	}

	node.RawLink = "ss://" + raw
	return node, nil
}

func isConfigURL(s string) bool {
	s = strings.TrimSpace(s)
	return strings.HasPrefix(s, "vmess://") ||
		strings.HasPrefix(s, "vless://") ||
		strings.HasPrefix(s, "trojan://") ||
		strings.HasPrefix(s, "ss://")
}

func (svc *SubscriptionService) isBase64Content(content string) bool {
	content = strings.TrimSpace(content)
	if len(content) < 10 {
		return false
	}
	_, err := base64.RawStdEncoding.DecodeString(content)
	if err != nil {
		_, err = base64.StdEncoding.DecodeString(content)
	}
	return err == nil
}

func decodeBase64Safe(s string) (string, error) {
	pad := 4 - len(s)%4
	if pad != 4 {
		for i := 0; i < pad; i++ {
			s += "="
		}
	}
	b, err := base64.URLEncoding.DecodeString(s)
	if err != nil {
		b, err = base64.StdEncoding.DecodeString(s)
	}
	if err != nil {
		b, err = base64.RawURLEncoding.DecodeString(s)
	}
	if err != nil {
		b, err = base64.RawStdEncoding.DecodeString(s)
	}
	return string(b), err
}

func unescapeName(s string) string {
	decoded, err := url.QueryUnescape(s)
	if err != nil {
		return s
	}
	return decoded
}

func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func getInt(m map[string]interface{}, key string) int {
	if v, ok := m[key]; ok {
		switch n := v.(type) {
		case float64:
			return int(n)
		case int:
			return n
		}
	}
	return 0
}
