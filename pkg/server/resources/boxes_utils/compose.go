package boxes_utils

import (
	"context"
	"fmt"
	"strings"

	"github.com/danielgtaylor/huma/v2"
	"github.com/dboxed/dboxed/pkg/server/db/dmodel"
	querier2 "github.com/dboxed/dboxed/pkg/server/db/querier"
	"github.com/dboxed/dboxed/pkg/server/models"
	"github.com/dboxed/dboxed/pkg/util"
)

func CreateComposeProject(c context.Context, box *dmodel.Box, req models.CreateBoxComposeProject) error {
	q := querier2.GetQuerier(c)

	err := util.CheckName(req.Name)
	if err != nil {
		return err
	}
	if strings.HasPrefix(req.Name, "dboxed-") {
		return huma.Error400BadRequest("'dboxed-' is a reserved internal prefix and can't be used")
	}

	bcps, err := dmodel.ListBoxComposeProjects(q, box.ID)
	if err != nil {
		return err
	}
	for _, bcp := range bcps {
		if bcp.Name == req.Name {
			return huma.Error400BadRequest(fmt.Sprintf("compose project with name %s already exists", req.Name))
		}
	}

	cp := dmodel.BoxComposeProject{
		BoxID:          box.ID,
		Name:           req.Name,
		ComposeProject: req.ComposeProject,
	}
	err = cp.Create(q)
	if err != nil {
		return err
	}

	err = ValidateBoxSpec(c, box, false)
	if err != nil {
		return err
	}

	return nil
}

func UpdateComposeProject(c context.Context, box *dmodel.Box, composeName string, composeProject string) error {
	q := querier2.GetQuerier(c)

	cp, err := dmodel.GetBoxComposeProjectByName(q, box.ID, composeName)
	if err != nil {
		return err
	}

	err = cp.UpdateComposeProject(q, composeProject)
	if err != nil {
		return err
	}

	err = ValidateBoxSpec(c, box, false)
	if err != nil {
		return err
	}

	err = dmodel.BumpChangeSeq(q, box)
	if err != nil {
		return err
	}

	return nil
}

func DeleteComposeProject(c context.Context, box *dmodel.Box, composeName string) error {
	q := querier2.GetQuerier(c)

	cp, err := dmodel.GetBoxComposeProjectByName(q, box.ID, composeName)
	if err != nil {
		return err
	}

	err = querier2.DeleteOneByFields[dmodel.BoxComposeProject](q, map[string]any{
		"box_id": box.ID,
		"name":   cp.Name,
	})
	if err != nil {
		return err
	}

	err = ValidateBoxSpec(c, box, false)
	if err != nil {
		return err
	}

	err = dmodel.BumpChangeSeq(q, box)
	if err != nil {
		return err
	}

	return nil
}
