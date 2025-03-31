package util

import (
	"go-fission-activity/config/initConfig"
	"strings"
)

// ArrayStringContains 判断数组中是否包含某个元素
func ArrayStringContains(arr []string, elem string) bool {
	for _, v := range arr {
		if v == elem {
			return true
		}
	}
	return false
}

// StartsWithPrefix 判断字符串中是否包含某个前缀
func StartsWithPrefix(str string, prefixes []string) bool {
	if initConfig.IsConfigActivity() {
		for _, prefix := range prefixes {
			if strings.HasPrefix(str, prefix) {
				return true
			}
		}
		return false
	}
	return true
}
