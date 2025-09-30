package flags

import "github.com/dboxed/dboxed/pkg/baseclient"

type GlobalFlags struct {
	Debug bool `help:"Enable debugging mode"`

	ClientAuthFile *string `help:"Override client auth file. Defaults to ~/.dboxed/client-auto.yaml" type:"path"`

	WorkDir string `help:"dboxed work dir" default:"/var/lib/dboxed"`
}

func (f *GlobalFlags) BuildClient() (*baseclient.Client, error) {
	return baseclient.FromClientAuthFile(f.ClientAuthFile)
}
