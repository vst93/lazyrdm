package service

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// 禁用鼠标输入
func disableMouseInput() {
	if runtime.GOOS == "windows" {
		return
	}
	// 禁用所有鼠标报告模式
	fmt.Print("\033[?1000l") // 禁用常规鼠标点击
	fmt.Print("\033[?1002l") // 禁用鼠标拖动
	fmt.Print("\033[?1003l") // 禁用所有鼠标事件
	fmt.Print("\033[?1006l") // 禁用SGR扩展鼠标模式
}

// 启用鼠标输入
func enableMouseInput() {
	if runtime.GOOS == "windows" {
		return
	}
	// 启用鼠标支持
	fmt.Print("\033[?1000h") // 启用常规鼠标点击
	fmt.Print("\033[?1002h") // 启用鼠标拖动
	fmt.Print("\033[?1003h") // 启用所有鼠标事件
	fmt.Print("\033[?1006h") // 启用SGR扩展鼠标模式
}

// 完全重置终端
func resetTerminalCompletely() {
	// 1. 使用 stty 重置终端（Unix-like 系统）
	if runtime.GOOS != "windows" {
		cmd := exec.Command("stty", "sane")
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Run()
	}

	// 2. 发送终端重置序列
	fmt.Print("\033c")     // 完全重置终端
	fmt.Print("\033[2J")   // 清屏
	fmt.Print("\033[H")    // 光标回到左上角
	fmt.Print("\033[?25h") // 显示光标
}

// 获取对终端友好的编辑器
func getTerminalFriendlyEditor() (string, []string, error) {
	// 检查环境变量
	if visual := os.Getenv("VISUAL"); visual != "" {
		return parseEditorCommand(visual)
	}

	if editor := os.Getenv("EDITOR"); editor != "" {
		return parseEditorCommand(editor)
	}

	// 平台特定的默认编辑器
	if runtime.GOOS == "windows" {
		// Windows 上使用记事本
		return "notepad", nil, nil
	} else {
		// Unix-like 系统上，优先使用对终端友好的编辑器
		for _, editor := range []string{"vim", "vi", "nano", "micro"} {
			if path, err := exec.LookPath(editor); err == nil {
				return path, nil, nil
			}
		}
		return "", nil, fmt.Errorf("未找到可用的编辑器")
	}
}

// 修改原始的 EditWithExternalEditor 函数，使用终端友好的编辑器
func EditWithExternalEditor(initialText string) (string, error) {
	// 创建临时文件
	tmpfile, err := os.CreateTemp("", "gocui_edit_*.txt")
	if err != nil {
		return "", fmt.Errorf("创建临时文件失败: %v", err)
	}
	defer os.Remove(tmpfile.Name())

	tmpPath := tmpfile.Name()

	// 写入初始文本
	if _, err := tmpfile.WriteString(initialText); err != nil {
		return "", fmt.Errorf("写入临时文件失败: %v", err)
	}
	tmpfile.Close()

	// 获取编辑器
	editor, editorArgs, err := getTerminalFriendlyEditor()
	if err != nil {
		return "", err
	}

	// 构建命令
	var cmd *exec.Cmd
	if len(editorArgs) > 0 {
		args := append(editorArgs, tmpPath)
		cmd = exec.Command(editor, args...)
	} else {
		cmd = exec.Command(editor, tmpPath)
	}

	// 设置标准输入输出
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// 运行编辑器
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("编辑器执行错误: %v", err)
	}

	// 读取编辑后的内容
	content, err := os.ReadFile(tmpPath)
	if err != nil {
		return "", fmt.Errorf("读取临时文件失败: %v", err)
	}

	return string(content), nil
}

// parseEditorCommand 解析编辑器命令（支持带参数的命令）
func parseEditorCommand(editorCmd string) (string, []string, error) {
	// 简单的命令行解析，支持带引号的参数
	parts := splitCommand(editorCmd)
	if len(parts) == 0 {
		return "", nil, fmt.Errorf("编辑器命令为空")
	}

	editor := parts[0]
	args := parts[1:]

	// 检查编辑器是否存在
	if _, err := exec.LookPath(editor); err != nil {
		return "", nil, fmt.Errorf("找不到编辑器 '%s': %v", editor, err)
	}

	return editor, args, nil
}

// splitCommand 简单的命令行分割函数，支持引号
func splitCommand(line string) []string {
	var args []string
	var current strings.Builder
	inQuotes := false
	quoteChar := byte(' ')

	for i := 0; i < len(line); i++ {
		c := line[i]

		switch {
		case c == '"' || c == '\'':
			if !inQuotes {
				inQuotes = true
				quoteChar = c
			} else if c == quoteChar {
				inQuotes = false
				quoteChar = ' '
			} else {
				current.WriteByte(c)
			}
		case c == ' ' && !inQuotes:
			if current.Len() > 0 {
				args = append(args, current.String())
				current.Reset()
			}
		default:
			current.WriteByte(c)
		}
	}

	if current.Len() > 0 {
		args = append(args, current.String())
	}

	return args
}
