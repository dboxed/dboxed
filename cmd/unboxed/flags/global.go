package flags

type GlobalFlags struct {
	WorkDir string `help:"unboxed work dir" default:"/var/lib/unboxed"`
}
