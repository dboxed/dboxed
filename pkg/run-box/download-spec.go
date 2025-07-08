package run_box

import (
	"context"
	"fmt"
	"github.com/dustin/go-humanize"
	"github.com/koobox/unboxed/pkg/types"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sigs.k8s.io/yaml"
)

func (rn *RunBox) retrieveBoxSpec(ctx context.Context) (*types.BoxSpec, error) {
	u, err := url.Parse(rn.BoxUrl)
	if err != nil {
		return nil, err
	}

	var b []byte
	if u.Scheme == "file" {
		b, err = os.ReadFile(filepath.FromSlash(u.Path))
		if err != nil {
			return nil, err
		}
	} else if u.Scheme == "http" || u.Scheme == "https" {
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

func (rn *RunBox) retrieveBoxSpecHttp(ctx context.Context) ([]byte, error) {
	req, err := http.NewRequest("GET", rn.BoxUrl, nil)
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
