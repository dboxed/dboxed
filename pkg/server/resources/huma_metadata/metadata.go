package huma_metadata

import (
	"github.com/danielgtaylor/huma/v2"
	"github.com/dboxed/dboxed/pkg/server/huma_utils"
)

const SkipAuth = "skip-auth"
const NeedAdmin = "need-admin"
const NoToken = "no-token"

const SkipWorkspace = "skip-workspace"

func NeedAdminModifier() func(o *huma.Operation) {
	return huma_utils.MetadataModifier(NeedAdmin, true)
}

func NoTokenModifier() func(o *huma.Operation) {
	return huma_utils.MetadataModifier(NoToken, true)
}
