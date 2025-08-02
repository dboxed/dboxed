package start_box

import (
	"context"
	"fmt"
	"github.com/dboxed/dboxed/pkg/types"
	"github.com/dustin/go-humanize"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nkeys"
	"io"
	"log/slog"
	"net/http"
	"os"
	"sigs.k8s.io/yaml"
	"strings"
	"time"
)

func (rn *StartBox) retrieveBoxSpec(ctx context.Context) (*types.BoxSpec, error) {
	var err error
	var b []byte
	if rn.BoxUrl.Scheme == "file" {
		path := strings.TrimPrefix(rn.BoxUrl.String(), "file://")
		b, err = os.ReadFile(path)
		if err != nil {
			return nil, err
		}
	} else if rn.BoxUrl.Scheme == "http" || rn.BoxUrl.Scheme == "https" {
		b, err = rn.retrieveBoxSpecHttp(ctx)
		if err != nil {
			return nil, err
		}
	} else if rn.BoxUrl.Scheme == "nats" {
		b, err = rn.retrieveBoxSpecNats(ctx)
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

func (rn *StartBox) retrieveBoxSpecNats(ctx context.Context) ([]byte, error) {
	subject := rn.BoxUrl.Query().Get("subject")
	if subject == "" {
		return nil, fmt.Errorf("missing subject in nats url")
	}

	nkeySeed, err := os.ReadFile(rn.Nkey)
	if err != nil {
		return nil, err
	}
	kp, err := nkeys.FromSeed(nkeySeed)
	if err != nil {
		return nil, err
	}
	nkey, err := kp.PublicKey()
	if err != nil {
		return nil, err
	}
	slog.InfoContext(ctx, "connecting to nats",
		slog.Any("url", rn.BoxUrl.String()),
		slog.Any("nkey", nkey),
	)
	nc, err := nats.Connect(rn.BoxUrl.String(), nats.Nkey(nkey, kp.Sign))
	if err != nil {
		return nil, err
	}
	defer nc.Close()

	slog.InfoContext(ctx, "requesting box-spec", slog.Any("subject", subject))
	resp, err := nc.Request(subject, nil, time.Second*15)
	if err != nil {
		return nil, err
	}

	return resp.Data, nil
}
