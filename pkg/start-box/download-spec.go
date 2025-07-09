package start_box

import (
	"context"
	"fmt"
	"github.com/dustin/go-humanize"
	"github.com/koobox/unboxed/pkg/types"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sigs.k8s.io/yaml"
)

func (rn *StartBox) retrieveBoxSpec(ctx context.Context) (*types.BoxSpec, error) {
	var err error
	var b []byte
	if rn.BoxUrl.Scheme == "file" {
		b, err = os.ReadFile(filepath.FromSlash(rn.BoxUrl.Path))
		if err != nil {
			return nil, err
		}
	} else if rn.BoxUrl.Scheme == "http" || rn.BoxUrl.Scheme == "https" {
		b, err = rn.retrieveBoxSpecHttp(ctx)
		if err != nil {
			return nil, err
		}
	}

	var boxFile types.BoxFile
	err = yaml.Unmarshal(b, &boxFile)
	if err != nil {
		return nil, err
	}

	return &boxFile.Spec, nil
}

func (rn *StartBox) retrieveBoxSpecHttp(ctx context.Context) ([]byte, error) {
	req, err := http.NewRequest("GET", rn.BoxUrl.String(), nil)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(io.LimitReader(resp.Body, humanize.MByte))
	if err != nil {
		return nil, err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("failed to download box spec: status=%d, body=%s", resp.StatusCode, string(b))
	}

	return b, nil
}
