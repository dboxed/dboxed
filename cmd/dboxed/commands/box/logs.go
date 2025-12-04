package box

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/charmbracelet/huh"
	"github.com/dboxed/dboxed/cmd/dboxed/commands/commandutils"
	"github.com/dboxed/dboxed/cmd/dboxed/flags"
	"github.com/dboxed/dboxed/pkg/boxspec"
	"github.com/dboxed/dboxed/pkg/clients"
	"github.com/dboxed/dboxed/pkg/server/models"
)

type LogsCmd struct {
	Box   string  `help:"Box ID or name" required:"" arg:""`
	LogId *string `help:"Log ID to stream"`
	Since string  `help:"Start streaming from this time (duration like '5m' or RFC3339 timestamp)" default:"1h"`
}

func (cmd *LogsCmd) Run(g *flags.GlobalFlags) error {
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

	if cmd.LogId == nil {
		logs, err := c2.ListLogs(ctx, "box", box.ID)
		if err != nil {
			return err
		}

		if len(logs) == 0 {
			slog.Info("no logs found for box", slog.Any("box", box.Name))
			return nil
		}

		options := make([]huh.Option[string], len(logs))
		for i, l := range logs {
			label := fmt.Sprintf("%s (ID: %s, T: %s)", l.FileName, l.ID, l.LastLogTime)
			options[i] = huh.NewOption(label, l.ID)
		}

		var selectedLogID string

		form := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title(fmt.Sprintf("Select a log file for box '%s'", box.Name)).
					Options(options...).
					Value(&selectedLogID),
			),
		)

		err = form.Run()
		if err != nil {
			return err
		}

		cmd.LogId = &selectedLogID
	}

	if cmd.LogId == nil {
		return fmt.Errorf("no log ID specified")
	}

	var metadata models.LogMetadataModel
	formatLine := func(line boxspec.LogsLine) string {
		_ = metadata
		return fmt.Sprintf("%s %s\n", line.Time.String(), line.Line)
	}

	err = c2.StreamLogs(ctx, *cmd.LogId, cmd.Since, func(event interface{}) error {
		switch v := event.(type) {
		case models.LogMetadataModel:
			slog.Info("streaming logs",
				slog.Any("file", v.FileName),
				slog.Any("format", v.Format),
			)
			metadata = v
		case boxspec.LogsBatch:
			for _, line := range v.Lines {
				fmt.Print(formatLine(line))
			}
		default:
			// End of history marker or other events
			// Silently continue streaming
		}
		return nil
	})

	if err != nil {
		return err
	}

	return nil
}
