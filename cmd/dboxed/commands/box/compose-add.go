package box

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/dboxed/dboxed/cmd/dboxed/commands/commandutils"
	"github.com/dboxed/dboxed/cmd/dboxed/flags"
	"github.com/dboxed/dboxed/pkg/clients"
	"github.com/dboxed/dboxed/pkg/server/models"
	"github.com/dboxed/dboxed/pkg/util"
)

type AddComposeCmd struct {
	Box         string `help:"Specify the box" required:"" arg:""`
	ComposeFile string `help:"Path to docker-compose.yml file" required:"" short:"f"`
}

func (cmd *AddComposeCmd) Run(g *flags.GlobalFlags) error {
	ctx := context.Background()

	c, err := g.BuildClient(ctx)
	if err != nil {
		return err
	}

	b, err := commandutils.GetBox(ctx, c, cmd.Box)
	if err != nil {
		return err
	}

	name, content, err := LoadComposeFileForBox(cmd.ComposeFile)
	if err != nil {
		return err
	}

	c2 := &clients.BoxClient{Client: c}

	req := models.CreateBoxComposeProject{
		Name:           name,
		ComposeProject: string(content),
	}

	err = c2.CreateComposeProject(ctx, b.ID, req)
	if err != nil {
		return err
	}

	slog.Info("compose project created", slog.Any("box_id", b.ID), slog.Any("name", name))

	return nil
}

func LoadComposeFileForBox(nameAndPath string) (string, []byte, error) {
	s := strings.SplitN(nameAndPath, "=", 2)

	var name, path string
	if len(s) == 0 {
		return "", nil, fmt.Errorf("invalid --compose-file flag")
	}
	if len(s) == 2 {
		name = s[0]
		path = s[1]
	} else if len(s) == 1 {
		path = s[0]
	}

	y, content, err := util.UnmarshalYamlFileWithBytes[map[string]any](path)
	if err != nil {
		return "", nil, err
	}

	if name == "" {
		x, ok := (*y)["name"]
		if !ok {
			return "", nil, fmt.Errorf("could not determine compose project name. Either specifiy it in the form of '--compose-file <name>=<path>' or put it into the compose file itself, via the top-level 'name' field")
		}
		name, ok = x.(string)
		if !ok {
			return "", nil, fmt.Errorf("name in compose file is not a string")
		}
	}

	return name, content, nil
}
