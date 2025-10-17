package util

import (
	"regexp"
	"strings"

	"github.com/danielgtaylor/huma/v2"
)

const nameFmt string = "[a-z0-9]([-_a-z0-9]*[_a-z0-9])?"

var nameFmtRegex = regexp.MustCompile("^" + nameFmt + "$")

const nameMaxLen int = 63

func CheckName(name string, extraAllowedChars ...rune) error {
	if len(name) == 0 {
		return huma.Error400BadRequest("empty names not allowed")
	}
	if len(name) > nameMaxLen {
		return huma.Error400BadRequest("name is too long")
	}
	for _, c := range extraAllowedChars {
		name = strings.ReplaceAll(name, string(c), "")
	}
	if !nameFmtRegex.MatchString(name) {
		return huma.Error400BadRequest("name contains invalid characters")
	}
	return nil
}
