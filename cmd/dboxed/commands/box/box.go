package box

import (
	"github.com/dboxed/dboxed/cmd/dboxed/commands/box/volume"
)

type BoxCommands struct {
	Create   CreateCmd        `cmd:"" help:"Create a box"`
	Get      GetCmd           `cmd:"" help:"Get a box"`
	List     ListCmd          `cmd:"" help:"List boxes" aliases:"ls"`
	Delete   DeleteCmd        `cmd:"" help:"Delete a box" aliases:"rm,delete"`
	Start    StartCmd         `cmd:"" help:"Start a box"`
	Stop     StopCmd          `cmd:"" help:"Stop a box"`
	Status   StatusCmd        `cmd:"" help:"Display box run status and containers"`
	Logs     LogsCmd          `cmd:"" help:"Stream box logs"`
	ListLogs ListLogsCmd      `cmd:"" help:"List available log files for a box"`
	Volume   volume.VolumeCmd `cmd:"" help:"Manage volume attachments"`

	AddCompose    AddComposeCmd    `cmd:"" help:"Create a compose project" group:"compose"`
	RemoveCompose RemoveComposeCmd `cmd:"" help:"Remove a compose project" group:"compose" aliases:"rm-compose,delete-compose"`
	ListCompose   ListComposeCmd   `cmd:"" help:"List compose projects" group:"compose" aliases:"ls-compose"`
	UpdateCompose UpdateComposeCmd `cmd:"" help:"Update a compose project" group:"compose"`
}
