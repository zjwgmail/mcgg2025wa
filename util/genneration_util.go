package util

import (
	"fmt"
	"strconv"
)

func GetNewGeneration(generation string) (string, error) {
	num, err := strconv.Atoi(generation)
	if err != nil {
		return "", err
	}
	if num < 8 {
		num = num + 1
	} else {
		num = 9
	}
	// 将数字 转换为两位字符串，左侧补零
	formattedStr := fmt.Sprintf("%02d", num)

	return formattedStr, nil
}
