package util

import (
	"fmt"
	"regexp"
	"strings"
)

const nameFmt string = "[a-z0-9]([-_a-z0-9]*[_a-z0-9])?"

var nameFmtRegex = regexp.MustCompile("^" + nameFmt + "$")

const nameMaxLen int = 63

func CheckName(name string, extraAllowedChars ...rune) error {
	if len(name) == 0 {
		return fmt.Errorf("empty names not allowed")
	}
	if len(name) > nameMaxLen {
		return fmt.Errorf("name is too long")
	}
	for _, c := range extraAllowedChars {
		name = strings.ReplaceAll(name, string(c), "")
	}
	if !nameFmtRegex.MatchString(name) {
		return fmt.Errorf("name contains invalid characters")
	}
	return nil
}
