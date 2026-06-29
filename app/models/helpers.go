package models

import (
	"encoding/base64"
	"net/url"
)

func EncodeBase64NoPadding(s string) string {
	return base64.RawURLEncoding.EncodeToString([]byte(s))
}

func encodeBase64(s string) string {
	return base64.StdEncoding.EncodeToString([]byte(s))
}

func urlEncode(s string) string {
	return url.QueryEscape(s)
}
