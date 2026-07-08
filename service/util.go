package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"unicode"
)

func PrintLn(str any) {
	_ = str
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

// DisposeMultibyteString 处理多字节字符 — DEPRECATED/UNUSED, kept for potential future use
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

// // OpenFileManager 使用系统默认文件管理器打开指定路径
// func OpenFileManager(path string) error {
// 	// 获取绝对路径
// 	absPath, err := filepath.Abs(path)
// 	if err != nil {
// 		return err
// 	}

// 	// 根据不同操作系统执行不同命令
// 	switch runtime.GOOS {
// 	case "darwin": // macOS
// 		return exec.Command("open", absPath).Start()
// 	case "windows": // Windows
// 		// 转换路径分隔符为Windows格式
// 		winPath := filepath.ToSlash(absPath)
// 		// 处理Windows驱动器号
// 		if len(winPath) >= 2 && winPath[1] == ':' {
// 			winPath = strings.ToUpper(string(winPath[0])) + winPath[1:]
// 		}
// 		return exec.Command("explorer", winPath).Start()
// 	default: // Linux和其他类Unix系统
// 		return exec.Command("xdg-open", absPath).Start()
// 	}
// }

// OpenFileManager 使用默认文件管理器打开文件所在目录
func OpenFileManager(filePath string) error {
	// 获取文件的绝对路径和所在目录
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return fmt.Errorf("获取绝对路径失败: %w", err)
	}

	dirPath := filepath.Dir(absPath)

	// 根据操作系统执行不同命令
	switch runtime.GOOS {
	case "darwin": // macOS
		// 使用 open 命令打开目录
		err = exec.Command("open", dirPath).Start()
	case "linux":
		// 使用 xdg-open 打开目录
		err = exec.Command("xdg-open", dirPath).Start()
	case "windows":
		// Windows 需要特殊处理路径格式
		winDir := strings.ReplaceAll(dirPath, "/", "\\")

		// 使用 PowerShell 命令打开目录
		cmd := exec.Command("powershell", "-Command",
			fmt.Sprintf("Start-Process explorer -ArgumentList '%s'", winDir))

		// 直接运行命令，不处理窗口隐藏
		err = cmd.Start()
	default:
		err = fmt.Errorf("不支持的操作系统: %s", runtime.GOOS)
	}

	return err
}

// GetDownloadPath 获取当前系统的默认下载目录路径
func GetDownloadPath() (string, error) {
	switch runtime.GOOS {
	case "windows":
		return getWindowsDownloadPath()
	case "darwin":
		return getMacDownloadPath()
	default: // Linux 和其他类Unix系统
		return getLinuxDownloadPath()
	}
}

// getWindowsDownloadPath 获取Windows下载路径
func getWindowsDownloadPath() (string, error) {
	// 首选检查环境变量
	if path := os.Getenv("USERPROFILE"); path != "" {
		return filepath.Join(path, "Downloads"), nil
	}

	// 备选方案：使用已知文件夹ID (FOLDERID_Downloads)
	// 这需要调用Windows API，简单实现如下：
	path, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}
	return filepath.Join(path, "Downloads"), nil
}

// getMacDownloadPath 获取macOS下载路径
func getMacDownloadPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}
	return filepath.Join(home, "Downloads"), nil
}

// getLinuxDownloadPath 获取Linux下载路径
func getLinuxDownloadPath() (string, error) {
	// 1. 检查XDG规范的环境变量
	if path := os.Getenv("XDG_DOWNLOAD_DIR"); path != "" {
		return path, nil
	}

	// 2. 检查用户目录下的Downloads
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}
	downloadsPath := filepath.Join(home, "Downloads")

	// 3. 检查路径是否存在（有些系统可能使用不同名称）
	if _, err := os.Stat(downloadsPath); err == nil {
		return downloadsPath, nil
	}

	// 4. 最后尝试使用$HOME作为备选
	return home, nil
}

func UnicodeSequenceToString(unicodeSeq string) (string, error) {
	var result strings.Builder

	// 处理 \uXXXX 格式
	for i := 0; i < len(unicodeSeq); {
		if i+6 <= len(unicodeSeq) && unicodeSeq[i:i+2] == "\\u" {
			// 提取十六进制部分
			hexStr := unicodeSeq[i+2 : i+6]
			code, err := strconv.ParseInt(hexStr, 16, 32)
			if err != nil {
				return "", err
			}
			result.WriteRune(rune(code))
			i += 6
		} else {
			result.WriteByte(unicodeSeq[i])
			i++
		}
	}

	return result.String(), nil
}

// 计算字符串的占位长度（中文2，英文1）
func DisplayWidth(s string) int {
	width := 0
	for _, r := range s {
		if unicode.In(r, unicode.Han) ||
			r >= 0xFF00 && r <= 0xFFEF || // 全角字符
			r >= 0x3000 && r <= 0x303F { // 中文标点
			width += 2
		} else {
			width += 1
		}
	}
	return width
}

func GetUserAgent() string {
	// 根据实际平台类型生成模拟的浏览器 User-Agent
	switch runtime.GOOS {
	case "darwin":
		return "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.114 Safari/537.36"
	case "windows":
		return "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.114 Safari/537.36"
	default: // Linux 和其他类Unix系统
		return "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.114 Safari/537.36"
	}
}

func PostJson(url string, msg []byte, headers map[string]string) (string, error) {
	client := &http.Client{}
	req, err := http.NewRequest("POST", url, strings.NewReader(string(msg)))
	if err != nil {
		return "", err
	}
	for key, header := range headers {
		req.Header.Set(key, header)
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

// SendAppStats 发送统计信息到 umami.dev
func SendAppStats() {
	website := "32c24ade-d689-4252-a37a-52c61aa04e5a"
	title := "lazyrdm"
	jsonMap := map[string]interface{}{
		"type": "event",
		"payload": map[string]interface{}{
			"website":  website,
			"screen":   "",
			"language": "",
			"title":    title,
			"hostname": "meimingzi.top",
			"url":      "https://meimingzi.top/" + title,
			"referrer": "",
		},
	}
	jsonStr, _ := json.Marshal(jsonMap)
	headers := map[string]string{
		"Content-Type": "application/json",
		"User-Agent":   GetUserAgent(),
	}
	PostJson("https://api-gateway.umami.dev/api/send", jsonStr, headers)
}
