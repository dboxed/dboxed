package box

import (
	"context"
	"fmt"
	"os"

	"github.com/dboxed/dboxed/cmd/dboxed/commands/commandutils"
	"github.com/dboxed/dboxed/cmd/dboxed/flags"
	"github.com/dboxed/dboxed/pkg/clients"
	"sigs.k8s.io/yaml"
)

type ListComposeCmd struct {
	Box string `help:"Box ID, UUID, or name" required:"" arg:""`
}

type PrintCompose struct {
	Name     string `col:"Name"`
	Services string `col:"Services"`
}

func (cmd *ListComposeCmd) Run(g *flags.GlobalFlags) error {
	ctx := context.Background()

	c, err := g.BuildClient(ctx)
	if err != nil {
		return err
	}

	b, err := commandutils.GetBox(ctx, c, cmd.Box)
	if err != nil {
		return err
	}

	c2 := &clients.BoxClient{Client: c}

	projects, err := c2.ListComposeProjects(ctx, b.ID)
	if err != nil {
		return err
	}

	var table []PrintCompose
	for _, cp := range projects {
		services := "<unknown>"

		var y map[string]any
		err = yaml.Unmarshal([]byte(cp.ComposeProject), &y)
		if err == nil {
			x, ok := y["services"]
			if ok {
				x2, ok := x.(map[string]any)
				if ok {
					services = fmt.Sprintf("%d", len(x2))
				}
			}
		}

		table = append(table, PrintCompose{
			Name:     cp.Name,
			Services: services,
		})
	}

	err = commandutils.PrintTable(os.Stdout, table)
	if err != nil {
		return err
	}

	return nil
}
