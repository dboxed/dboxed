package flags

type GlobalFlags struct {
	Debug bool `help:"Enable debugging mode"`

	WorkDir string `help:"unboxed work dir" default:"/var/lib/unboxed"`
}
