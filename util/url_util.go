package util

import (
	"net/url"
)

// QueryEscape url加密
func QueryEscape(escapeUrlStr string) string {
	return url.QueryEscape(escapeUrlStr)
}

// QueryUnescape url解密
func QueryUnescape(escapeUrlStr string) (string, error) {
	return url.QueryUnescape(escapeUrlStr)
}
