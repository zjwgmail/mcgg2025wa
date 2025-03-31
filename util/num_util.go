package util

import "strconv"

// ToBase32 将数字转换为字符串
func ToBase32(num int) string {
	num = num + 1888888

	chars := "5dysjk0replh3tn2og4w7ca9bf6um8vx1iqz"
	result := ""
	for num > 0 {
		remainder := num % 36
		result = string(chars[remainder]) + result
		num /= 36
	}

	return result
}

// AddThousandSeparators 函数将整数转换为带有千分号的字符串
func AddThousandSeparators(number int) string {
	strNumber := strconv.Itoa(number)

	runes := []rune(strNumber)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	// 从右边开始，每三位添加一个逗号
	var result string
	index := 0
	for i := 1; i <= len(runes); i++ {
		if i%3 == 0 {
			if i == len(runes) {
				result += string(runes[i-3:])
			} else {
				result = result + string(runes[i-3:i]) + ","
			}
			index = i
		}
	}
	if index < len(runes) {
		result = result + string(runes[index:])
	}

	resultRunes := []rune(result)
	for i, j := 0, len(resultRunes)-1; i < j; i, j = i+1, j-1 {
		resultRunes[i], resultRunes[j] = resultRunes[j], resultRunes[i]
	}
	return string(resultRunes)
}

// AddThousandSeparators64 函数将整数转换为带有千分号的字符串
func AddThousandSeparators64(number int64) string {
	strNumber := strconv.FormatInt(number, 10)
	runes := []rune(strNumber)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	// 从右边开始，每三位添加一个逗号
	var result string
	index := 0
	for i := 1; i <= len(runes); i++ {
		if i%3 == 0 {
			if i == len(runes) {
				result += string(runes[i-3:])
			} else {
				result = result + string(runes[i-3:i]) + ","
			}
			index = i
		}
	}
	if index < len(runes) {
		result = result + string(runes[index:])
	}

	resultRunes := []rune(result)
	for i, j := 0, len(resultRunes)-1; i < j; i, j = i+1, j-1 {
		resultRunes[i], resultRunes[j] = resultRunes[j], resultRunes[i]
	}
	return string(resultRunes)
}
