package input

import (
	"bufio"
	"os"
	"strconv"
	"strings"

	"github.com/pterm/pterm"
)

func GetInput() string {
	buf, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil {
		panic(err)
	}
	return string(strings.TrimSpace(string(buf)))
}

func GetValidInput() string {
	for {
		s := GetInput()
		if s == "" {
			pterm.Error.Println("无效输入，输入不能为空")
			continue
		}
		return s
	}
}

func GetValidInt() int {
	for {
		s := GetInput()
		if s == "" {
			pterm.Error.Println("无效输入，输入不能为空")
			continue
		}
		v, err := strconv.Atoi(s)
		if err != nil {
			pterm.Error.Println("无效输入，输入必须为数字")
			continue
		}
		return v
	}
}

func GetValidIntOrEmpty() (int, bool) {
	for {
		s := GetInput()
		if s == "" {
			return 0, true
		}
		v, err := strconv.Atoi(s)
		if err != nil {
			pterm.Error.Println("无效输入，输入必须为数字")
			continue
		}
		return v, false
	}
}

func GetInputYN(defaultSelection bool) bool {
	s := GetInput()
	if strings.HasPrefix(s, "Y") || strings.HasPrefix(s, "y") || strings.TrimSpace(s) == "" {
		return true
	} else if strings.HasPrefix(s, "N") || strings.HasPrefix(s, "n") {
		return false
	}
	return defaultSelection
}
