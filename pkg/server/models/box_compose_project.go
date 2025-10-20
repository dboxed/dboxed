package models

import "github.com/dboxed/dboxed/pkg/server/db/dmodel"

type BoxComposeProject struct {
	Name           string `json:"name"`
	ComposeProject string `json:"composeProject"`
}

type CreateBoxComposeProject struct {
	Name           string `json:"name"`
	ComposeProject string `json:"composeProject"`
}

type UpdateBoxComposeProject struct {
	ComposeProject string `json:"composeProject"`
}

func BoxComposeProjectFromDB(s dmodel.BoxComposeProject) *BoxComposeProject {
	return &BoxComposeProject{
		Name:           s.Name,
		ComposeProject: s.ComposeProject,
	}
}
