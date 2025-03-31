package util

import (
	"crypto/md5"
	"encoding/hex"
	"strings"
)

// CallSign 计算签名
func CallSign(headers map[string]string, body string, accessSecret string) string {

	var raw strings.Builder
	raw.WriteString("accessKey=")
	raw.WriteString(headers["accessKey"])
	raw.WriteString("&action=")
	raw.WriteString(headers["action"])
	raw.WriteString("&bizType=")
	raw.WriteString(headers["bizType"])
	raw.WriteString("&ts=")
	raw.WriteString(headers["ts"])

	// 如果body不为空，则添加到raw中
	if body != "" {
		raw.WriteString("&body=")
		raw.WriteString(body)
	}

	// 添加accessSecret
	raw.WriteString("&accessSecret=")
	raw.WriteString(accessSecret)

	// 计算MD5哈希值
	hash := md5.Sum([]byte(raw.String()))
	return hex.EncodeToString(hash[:])
}

func CallSignFormData(headers map[string]string, accessSecret string) string {

	var raw strings.Builder
	raw.WriteString("accessKey=")
	raw.WriteString(headers["accessKey"])
	raw.WriteString("&action=")
	raw.WriteString(headers["action"])
	raw.WriteString("&bizType=")
	raw.WriteString(headers["bizType"])
	raw.WriteString("&ts=")
	raw.WriteString(headers["ts"])

	// 添加accessSecret
	raw.WriteString("&accessSecret=")
	raw.WriteString(accessSecret)

	// 计算MD5哈希值
	hash := md5.Sum([]byte(raw.String()))
	return hex.EncodeToString(hash[:])
}
