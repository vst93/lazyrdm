package service

import (
	"fmt"
	"strings"
)

func NewColorString(text string, parameters ...string) string {
	colorInt := 30
	bgColorInt := 40
	displayInt := 0

	color := ""
	bgColor := ""
	display := ""
	for i := range parameters {
		switch i {
		case 0:
			color = parameters[i]
		case 1:
			bgColor = parameters[i]
		case 2:
			display = parameters[i]
		default:
		}
	}

	// font color
	switch color {
	case "black":
		colorInt = 30
	case "red":
		colorInt = 31
	case "green":
		colorInt = 32
	case "yellow":
		colorInt = 33
	case "blue":
		colorInt = 34
	case "purple":
		colorInt = 35
	case "cyan":
		colorInt = 36
	case "white":
		colorInt = 37
	default:
		colorInt = 37
	}

	// background color
	switch bgColor {
	case "black":
		bgColorInt = 40
	case "red":
		bgColorInt = 41
	case "green":
		bgColorInt = 42
	case "yellow":
		bgColorInt = 43
	case "blue":
		bgColorInt = 44
	case "purple":
		bgColorInt = 45
	case "cyan":
		bgColorInt = 46
	case "white":
		bgColorInt = 47
	default:
		bgColorInt = 40
	}

	// display mode
	switch display {
	case "bold":
		displayInt = 1
	case "underline":
		displayInt = 4
	case "blink":
		displayInt = 5
	case "reverse":
		displayInt = 7
	case "conceal":
		displayInt = 8
	default:
		displayInt = 0
	}

	// fmt.Println(colorInt, bgColorInt, displayInt)

	return fmt.Sprintf("\x1b[%d;%d;%dm%s\x1b[0m", displayInt, colorInt, bgColorInt, text)
}

func NewTypeWord(str string, parameters ...string) string {
	str = strings.ToLower(str)
	isFull := false
	for _, v := range parameters {
		switch v {
		case "full":
			isFull = true
		default:
		}
	}

	ret := NewColorString("[Null]", "white", "red", "bold")
	switch str {
	case "string":
		// ret = "🇸"
		if isFull {
			ret = NewColorString("[String]", "white", "purple", "bold")
		} else {
			ret = NewColorString("[Str]", "white", "purple", "bold")
		}
	case "list":
		// ret = "🇱"
		ret = NewColorString("[List]", "black", "green", "bold")
	case "set":
		// ret = "🇪"
		ret = NewColorString("[Set]", "black", "yellow", "bold")
	case "zset":
		// ret = "🇿"
		ret = NewColorString("[ZSet]", "white", "red", "bold")
	case "hash":
		// ret = "🇭"
		ret = NewColorString("[Hash]", "black", "cyan", "bold")
	case "stream":
		// ret = "🇽"
		ret = NewColorString("[Stream]", "white", "red", "bold")
	case "json":
		// ret = "🇯"
		ret = NewColorString("[JSON]", "white", "yellow", "bold")
	default:
	}
	return ret
}
