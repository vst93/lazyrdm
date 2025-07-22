package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"unicode"
)

func PrintLn(str any) {
	// return
	//写入日志到文件
	f, _ := os.OpenFile("go_log.txt", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	defer f.Close()
	f.WriteString(fmt.Sprintln(str))
}

func PrettyString(str string) (string, error) {
	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, []byte(str), "", " "); err != nil {
		return str, err
	}
	return prettyJSON.String(), nil
}

func IsNormalChar(r rune) bool {
	const allowedSymbols = " _-.@,'[]{}()【】，？！:："
	if unicode.IsLetter(r) || unicode.IsNumber(r) {
		return true
	}
	// 检查字符是否在允许的符号字符串中
	for _, s := range allowedSymbols {
		if r == s {
			return true
		}
	}
	return false
}

// DisposeMultibyteString 处理多字节字符
func DisposeMultibyteString(text string) []byte {
	if len(text) == 0 {
		return []byte("")
	}
	var result []rune
	for _, r := range text {
		if r > 255 {
			result = append(result, r, 32)
		} else {
			result = append(result, r)
		}
	}
	return []byte(string(result))
}

func ToString(s any) string {
	switch s := s.(type) {
	case string:
		return s
	case []byte:
		return string(s)
	default:
		return fmt.Sprintf("%v", s)
	}
}
