package flags

import (
	"fmt"
	"net/url"
)

type BoxUrlFlags struct {
	BoxUrl  *url.URL `help:"Specify box url." required:"" xor:"box-url"`
	BoxFile string   `help:"Specify file." required:"" xor:"box-url" type:"existingfile"`
}

func (f *BoxUrlFlags) GetBoxUrl() (*url.URL, error) {
	if f.BoxUrl != nil {
		return f.BoxUrl, nil
	}

	x, err := url.Parse(fmt.Sprintf("file://%s", f.BoxFile))
	if err != nil {
		return nil, err
	}
	return x, nil
}
