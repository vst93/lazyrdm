package service

import (
	"fmt"
	"os"
	"os/exec"
)

// EditWithExternalEditor 使用外部编辑器编辑文本
func EditWithExternalEditor(initialText string) (string, error) {
	// 1. 创建临时文件
	tmpfile, err := os.CreateTemp("", "gocui_edit_*.txt")
	if err != nil {
		return "", fmt.Errorf("创建临时文件失败: %v", err)
	}
	defer os.Remove(tmpfile.Name()) // 确保函数返回前删除临时文件

	// 2. 将初始文本写入临时文件
	if _, err := tmpfile.WriteString(initialText); err != nil {
		return "", fmt.Errorf("写入临时文件失败: %v", err)
	}
	// 写入后立即关闭文件，确保编辑器能读写
	if err := tmpfile.Close(); err != nil {
		return "", fmt.Errorf("关闭临时文件失败: %v", err)
	}

	// 3. 确定并启动编辑器
	// 编辑器选择优先级可以参考：GIT_EDITOR > VISUAL > EDITOR > 默认vi [citation:3]
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vi" // 默认使用 vi
	}

	cmd := exec.Command(editor, tmpfile.Name())
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// 启动并等待编辑器退出
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("编辑器执行错误: %v", err)
	}

	// 4. 读取编辑后的内容
	content, err := os.ReadFile(tmpfile.Name())
	if err != nil {
		return "", fmt.Errorf("读取临时文件失败: %v", err)
	}

	return string(content), nil
}
