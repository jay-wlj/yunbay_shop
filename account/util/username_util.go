package util

import (
	"fmt"
	"regexp"
)

var usernameRegex *regexp.Regexp

func init() {
	var err error
	usernameRegex, err = regexp.Compile(`^[a-z0-9_]{3,32}$`)
	if err != nil {
		fmt.Printf("regexp compile error: %s", err)
		return
	}
}

// 检测用户名是否合法
func UsernameCheck(username string) (valid bool) {
	return usernameRegex.Match([]byte(username))
}
