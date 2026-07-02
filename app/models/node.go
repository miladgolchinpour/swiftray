package models

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
)

type NodeProtocol string

const (
	ProtocolVMess       NodeProtocol = "vmess"
	ProtocolVLess       NodeProtocol = "vless"
	ProtocolTrojan      NodeProtocol = "trojan"
	ProtocolShadowsocks NodeProtocol = "ss"
)

type NodeTransport string

const (
	TransportTCP          NodeTransport = "tcp"
	TransportWS           NodeTransport = "ws"
	TransportGRPC         NodeTransport = "grpc"
	TransportHTTPUpgrade  NodeTransport = "httpupgrade"
	TransportSplitHTTP    NodeTransport = "splithttp"
	TransportXHTTP        NodeTransport = "xhttp"
	TransportHTTP         NodeTransport = "h2"
)

type Node struct {
	ID               string         `json:"id"`
	Name             string         `json:"name"`
	Address          string         `json:"address"`
	Port             int            `json:"port"`
	ProtocolType     NodeProtocol   `json:"protocolType"`
	Transport        NodeTransport  `json:"transport"`
	TLS              bool           `json:"tls"`
	Delay            *int           `json:"delay"`
	UUID             string         `json:"uuid"`
	Password         string         `json:"password"`
	AlterID          int            `json:"alterId"`
	Security         string         `json:"security"`
	Cipher           string         `json:"cipher"`
	Flow             string         `json:"flow"`
	Encryption       string         `json:"encryption"`
	SNI              string         `json:"sni"`
	Fingerprint      string         `json:"fingerprint"`
	ALPN             string         `json:"alpn"`
	RealityPublicKey string         `json:"realityPublicKey"`
	RealityShortID   string         `json:"realityShortId"`
	RealitySpiderX   string         `json:"realitySpiderX"`
	Host             string         `json:"host"`
	Path             string         `json:"path"`
	ServiceName      string         `json:"serviceName"`
	ServiceMode      string         `json:"serviceMode"`
	RawLink          string         `json:"rawLink"`
}

func NewNode() Node {
	return Node{
		ID:             uuid.New().String(),
		ProtocolType:   ProtocolVLess,
		Transport:      TransportTCP,
		Port:           443,
		Security:       "auto",
		Encryption:     "none",
	}
}

func (n Node) DelayString() string {
	if n.Delay == nil {
		return "-- ms"
	}
	if *n.Delay == -1 {
		return "Fail"
	}
	return fmt.Sprintf("%d ms", *n.Delay)
}

func CompareDelay(a, b Node) bool {
	switch {
	case a.Delay == nil && b.Delay == nil:
		return false
	case a.Delay == nil:
		return false
	case b.Delay == nil:
		return true
	case *a.Delay == -1 && *b.Delay == -1:
		return false
	case *a.Delay == -1:
		return false
	case *b.Delay == -1:
		return true
	default:
		return *a.Delay < *b.Delay
	}
}

func (n Node) ConfigURL() string {
	switch n.ProtocolType {
	case ProtocolVMess:
		return n.toVMessURL()
	case ProtocolVLess:
		return n.toVLessURL()
	case ProtocolTrojan:
		return n.toTrojanURL()
	case ProtocolShadowsocks:
		return n.toShadowsocksURL()
	default:
		return ""
	}
}

func (n Node) toVMessURL() string {
	tlsStr := "insecure"
	if n.TLS {
		tlsStr = "tls"
	}

	jsonStr := fmt.Sprintf(
		`{"v":"2","ps":"%s","add":"%s","port":"%d","id":"%s","aid":"%d","scy":"%s","net":"%s","type":"none","host":"%s","path":"%s","tls":"%s","sni":"%s","fp":"%s"}`,
		n.Name, n.Address, n.Port, n.UUID, n.AlterID, n.SecurityOrAuto(), n.Transport, n.Host, n.Path, tlsStr, n.SNI, n.Fingerprint,
	)
	return "vmess://" + encodeBase64(jsonStr)
}

func (n Node) toVLessURL() string {
	params := []string{"type=" + string(n.Transport)}
	if n.TLS {
		params = append(params, "security=tls")
	}
	if n.SNI != "" {
		params = append(params, "sni="+n.SNI)
	}
	if n.Fingerprint != "" {
		params = append(params, "fp="+n.Fingerprint)
	}
	if n.Host != "" {
		params = append(params, "host="+n.Host)
	}
	if n.Path != "" {
		params = append(params, "path="+urlEncode(n.Path))
	}
	if n.Flow != "" {
		params = append(params, "flow="+n.Flow)
	}
	query := strings.Join(params, "&")
	name := urlEncode(n.Name)
	return fmt.Sprintf("vless://%s@%s:%d?%s#%s", n.UUID, n.Address, n.Port, query, name)
}

func (n Node) toTrojanURL() string {
	params := []string{"type=" + string(n.Transport)}
	if n.TLS {
		params = append(params, "security=tls")
	}
	if n.SNI != "" {
		params = append(params, "sni="+n.SNI)
	}
	if n.Fingerprint != "" {
		params = append(params, "fp="+n.Fingerprint)
	}
	if n.Host != "" {
		params = append(params, "host="+n.Host)
	}
	if n.Path != "" {
		params = append(params, "path="+urlEncode(n.Path))
	}
	query := strings.Join(params, "&")
	name := urlEncode(n.Name)
	return fmt.Sprintf("trojan://%s@%s:%d?%s#%s", n.Password, n.Address, n.Port, query, name)
}

func (n Node) toShadowsocksURL() string {
	cred := n.Cipher + ":" + n.Password
	b64 := EncodeBase64NoPadding(cred)
	name := urlEncode(n.Name)
	return fmt.Sprintf("ss://%s@%s:%d#%s", b64, n.Address, n.Port, name)
}

func (n Node) SecurityOrAuto() string {
	if n.Security != "" {
		return n.Security
	}
	return "auto"
}

func NormalizeTransport(t NodeTransport) NodeTransport {
	if t == TransportXHTTP {
		return TransportSplitHTTP
	}
	return t
}
