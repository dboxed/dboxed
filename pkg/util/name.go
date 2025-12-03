package util

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/danielgtaylor/huma/v2"
)

const nameFmt string = "[a-z0-9]([-_a-z0-9]*[_a-z0-9])?"

var nameFmtRegex = regexp.MustCompile("^" + nameFmt + "$")

const NameMaxLen int = 63

type CheckNameOptions struct {
	ExtraAllowedChars []rune
	MaxLen            int
}

func CheckName(name string) error {
	return CheckNameOpts(name, CheckNameOptions{})
}

func CheckNameOpts(name string, opts CheckNameOptions) error {
	maxLen := NameMaxLen
	if opts.MaxLen != 0 {
		maxLen = opts.MaxLen
	}

	if len(name) == 0 {
		return huma.Error400BadRequest("empty names not allowed")
	}
	if len(name) > maxLen {
		return huma.Error400BadRequest(fmt.Sprintf("name is longer then %d characters", maxLen))
	}
	if strings.ToLower(name) != name {
		return huma.Error400BadRequest("names can only contain lowercase characters")
	}
	for _, c := range opts.ExtraAllowedChars {
		name = strings.ReplaceAll(name, string(c), "")
	}
	if !nameFmtRegex.MatchString(name) {
		return huma.Error400BadRequest("name contains invalid characters")
	}
	return nil
}
