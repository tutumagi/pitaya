package utils

import (
	"unicode/utf8"
)

const (
	CodeStrLengthOK   = 1
	CodeStrInvalid    = 2
	CodeStrOverLength = 3
)

// 字符串是否超过长度，maxLength 是指英文的长度，str里面英文算一个字符，其他的算2个字符
func CheckLength(str string, maxLength int32) int {
	if !utf8.ValidString(str) {
		return CodeStrInvalid
	}
	length := 0
	for _, code := range str {
		if code == utf8.RuneError {
			return CodeStrInvalid
		}
		if code > utf8.MaxRune {
			return CodeStrInvalid
		}
		if code > 255 { // 如果不是 ASCII 字符
			length += 2
		} else {
			length++
		}
	}
	if length > int(maxLength) {
		return CodeStrOverLength
	}
	return CodeStrLengthOK
}
