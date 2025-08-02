package flags

type GlobalFlags struct {
	Debug bool `help:"Enable debugging mode"`

	WorkDir string `help:"dboxed work dir" default:"/var/lib/dboxed"`
}
