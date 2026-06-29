package models

import "fmt"

type IPInfo struct {
	IPVersion       *int     `json:"ipVersion"`
	IPAddress       string   `json:"ipAddress"`
	Latitude        float64  `json:"latitude"`
	Longitude       float64  `json:"longitude"`
	CountryName     string   `json:"countryName"`
	CountryCode     string   `json:"countryCode"`
	Capital         string   `json:"capital"`
	PhoneCodes      []int    `json:"phoneCodes"`
	TimeZones       []string `json:"timeZones"`
	ZipCode         string   `json:"zipCode"`
	CityName        string   `json:"cityName"`
	RegionName      string   `json:"regionName"`
	RegionCode      string   `json:"regionCode"`
	Continent       string   `json:"continent"`
	ContinentCode   string   `json:"continentCode"`
	Currencies      []string `json:"currencies"`
	Languages       []string `json:"languages"`
	ASN             string   `json:"asn"`
	ASNOrganization string   `json:"asnOrganization"`
	IsProxy         bool     `json:"isProxy"`
}

func (i IPInfo) CountryFlag() string {
	if len(i.CountryCode) != 2 {
		return ""
	}
	r1 := rune(i.CountryCode[0]) - 'A' + 0x1F1E6
	r2 := rune(i.CountryCode[1]) - 'A' + 0x1F1E6
	return string([]rune{r1, r2})
}

func (i IPInfo) Summary() string {
	name := i.CountryName
	if name == "" {
		name = "Unknown"
	}
	return fmt.Sprintf("%s %s", i.CountryFlag(), name)
}

type LogLevel string

const (
	LevelInfo    LogLevel = "info"
	LevelDebug   LogLevel = "debug"
	LevelWarning LogLevel = "warning"
	LevelError   LogLevel = "error"
)

type LogEntry struct {
	ID        string   `json:"id"`
	Level     LogLevel `json:"level"`
	Message   string   `json:"message"`
	Timestamp string   `json:"timestamp"`
}
