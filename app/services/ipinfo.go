package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"time"

	"SwiftRay/app/models"
)

type IPInfoService struct {
	client  *http.Client
	cached  *models.IPInfo
	onUpdate func(models.IPInfo)
}

func NewIPInfoService() *IPInfoService {
	return &IPInfoService{
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

func (s *IPInfoService) OnUpdate(fn func(models.IPInfo)) {
	s.onUpdate = fn
}

func (s *IPInfoService) FetchViaProxy(socksPort int) (*models.IPInfo, error) {
	url := "https://free.freeipapi.com/api/json"

	cmd := exec.Command("curl", "--socks5-hostname", fmt.Sprintf("127.0.0.1:%d", socksPort), "-s", "--max-time", "10", url)
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("curl failed: %w", err)
	}

	return s.parseResponse(out)
}

func (s *IPInfoService) FetchDirect() (*models.IPInfo, error) {
	url := "https://free.freeipapi.com/api/json"

	resp, err := s.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("fetch failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}

	return s.parseResponse(body)
}

func (s *IPInfoService) parseResponse(data []byte) (*models.IPInfo, error) {
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("parse json: %w", err)
	}

	info := &models.IPInfo{}

	if v, ok := raw["ipAddress"].(string); ok {
		info.IPAddress = v
	}
	if v, ok := raw["countryName"].(string); ok {
		info.CountryName = v
	}
	if v, ok := raw["countryCode"].(string); ok {
		info.CountryCode = v
	}
	if v, ok := raw["cityName"].(string); ok {
		info.CityName = v
	}
	if v, ok := raw["regionName"].(string); ok {
		info.RegionName = v
	}
	if v, ok := raw["latitude"].(float64); ok {
		info.Latitude = v
	}
	if v, ok := raw["longitude"].(float64); ok {
		info.Longitude = v
	}
	if v, ok := raw["asn"].(string); ok {
		info.ASN = v
	}
	if v, ok := raw["asnOrganization"].(string); ok {
		info.ASNOrganization = v
	}
	if v, ok := raw["isProxy"].(bool); ok {
		info.IsProxy = v
	}
	if v, ok := raw["capital"].(string); ok {
		info.Capital = v
	}
	if v, ok := raw["continent"].(string); ok {
		info.Continent = v
	}
	if v, ok := raw["continentCode"].(string); ok {
		info.ContinentCode = v
	}
	if v, ok := raw["regionCode"].(string); ok {
		info.RegionCode = v
	}
	if v, ok := raw["zipCode"].(string); ok {
		info.ZipCode = v
	}

	if v, ok := raw["ipVersion"].(float64); ok {
		iv := int(v)
		info.IPVersion = &iv
	}

	if v, ok := raw["phoneCodes"].([]interface{}); ok {
		codes := make([]int, 0, len(v))
		for _, c := range v {
			if f, ok := c.(float64); ok {
				codes = append(codes, int(f))
			}
		}
		info.PhoneCodes = codes
	}

	if v, ok := raw["timeZones"].([]interface{}); ok {
		zones := make([]string, 0, len(v))
		for _, z := range v {
			if s, ok := z.(string); ok {
				zones = append(zones, s)
			}
		}
		info.TimeZones = zones
	}

	if v, ok := raw["currencies"].([]interface{}); ok {
		currencies := make([]string, 0, len(v))
		for _, c := range v {
			if s, ok := c.(string); ok {
				currencies = append(currencies, s)
			}
		}
		info.Currencies = currencies
	}

	if v, ok := raw["languages"].([]interface{}); ok {
		languages := make([]string, 0, len(v))
		for _, l := range v {
			if s, ok := l.(string); ok {
				languages = append(languages, s)
			}
		}
		info.Languages = languages
	}

	s.cached = info
	if s.onUpdate != nil {
		s.onUpdate(*info)
	}

	return info, nil
}

func (s *IPInfoService) GetCached() *models.IPInfo {
	return s.cached
}

func (s *IPInfoService) FetchForConnection(isConnected bool, socksPort int) (*models.IPInfo, error) {
	if isConnected && socksPort > 0 {
		return s.FetchViaProxy(socksPort)
	}
	return s.FetchDirect()
}
