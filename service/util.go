package service

import (
	"fmt"
	"os"
)

func PrintLn(str any) {
	return
	//写入日志到文件
	f, _ := os.OpenFile("/Users/vst/Downloads/go_log.txt", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	defer f.Close()
	f.WriteString(fmt.Sprintln("%s", str))
}
