package box

type BoxCommands struct {
	Create   CreateCmd   `cmd:"" help:"Create a box"`
	Get      GetCmd      `cmd:"" help:"Get a box"`
	List     ListCmd     `cmd:"" help:"List boxes" aliases:"ls"`
	Delete   DeleteCmd   `cmd:"" help:"Delete a box" aliases:"rm,delete"`
	Start    StartCmd    `cmd:"" help:"Start a box"`
	Stop     StopCmd     `cmd:"" help:"Stop a box"`
	Status   StatusCmd   `cmd:"" help:"Display box run status and containers"`
	Logs     LogsCmd     `cmd:"" help:"Stream box logs"`
	ListLogs ListLogsCmd `cmd:"" help:"List available log files for a box"`

	AddCompose    AddComposeCmd    `cmd:"" help:"Create a compose project" group:"compose"`
	RemoveCompose RemoveComposeCmd `cmd:"" help:"Remove a compose project" group:"compose" aliases:"rm-compose,delete-compose"`
	ListCompose   ListComposeCmd   `cmd:"" help:"List compose projects" group:"compose" aliases:"ls-compose"`
	UpdateCompose UpdateComposeCmd `cmd:"" help:"Update a compose project" group:"compose"`

	AttachVolume AttachVolumeCmd `cmd:"" help:"Attach a volume to a box" group:"volume"`
	DetachVolume DetachVolumeCmd `cmd:"" help:"Detach a volume from a box" group:"volume"`
	UpdateUpdate UpdateVolumeCmd `cmd:"" help:"Update volume attachment settings" group:"volume"`
	ListVolume   ListVolumesCmd  `cmd:"" help:"List attached volumes" aliases:"ls-volumes" group:"volume"`
}
