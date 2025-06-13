package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
)

func PrintLn(str any) {
	// return
	//写入日志到文件
	f, _ := os.OpenFile("go_log.txt", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	defer f.Close()
	f.WriteString(fmt.Sprintln(str))
}

func 						PrettyString(str string) (string, error) {
	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, []byte(str), "", " "); err != nil {
		return str, err
	}
	return prettyJSON.String(), nil
}
