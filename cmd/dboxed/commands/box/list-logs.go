package box

import (
	"context"
	"os"

	"github.com/dboxed/dboxed/cmd/dboxed/commands/commandutils"
	"github.com/dboxed/dboxed/cmd/dboxed/flags"
	"github.com/dboxed/dboxed/pkg/clients"
)

type ListLogsCmd struct {
	Box string `help:"Box ID or name" required:"" arg:""`
	flags.ListFlags
}

type PrintListLogs struct {
	ID          string `col:"ID"`
	CreateAt    string `col:"Created At"`
	FileName    string `col:"File Name"`
	Format      string `col:"Format"`
	LastLogTime string `col:"Last Log"`
}

func (cmd *ListLogsCmd) Run(g *flags.GlobalFlags) error {
	ctx := context.Background()

	c, err := g.BuildClient(ctx)
	if err != nil {
		return err
	}

	box, err := commandutils.GetBox(ctx, c, cmd.Box)
	if err != nil {
		return err
	}

	c2 := &clients.LogsClient{Client: c}

	logs, err := c2.ListLogs(ctx, "box", box.ID)
	if err != nil {
		return err
	}

	var table []PrintListLogs
	for _, l := range logs {
		p := PrintListLogs{
			ID:       l.ID,
			CreateAt: l.CreatedAt.String(),
			FileName: l.FileName,
			Format:   l.Format,
		}
		if l.LastLogTime != nil {
			p.LastLogTime = l.LastLogTime.String()
		}
		table = append(table, p)
	}

	err = commandutils.PrintTable(os.Stdout, table, cmd.ShowIds)
	if err != nil {
		return err
	}

	return nil
}
