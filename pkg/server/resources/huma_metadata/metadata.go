package huma_metadata

import (
	"github.com/danielgtaylor/huma/v2"
	"github.com/dboxed/dboxed/pkg/server/huma_utils"
)

const SkipAuth = "skip-auth"
const NeedAdmin = "need-admin"

const AllowWorkspaceToken = "allow-workspace-token"
const AllowBoxToken = "allow-box-token"

const SkipWorkspace = "skip-workspace"

func NeedAdminModifier() func(o *huma.Operation) {
	return huma_utils.MetadataModifier(NeedAdmin, true)
}
