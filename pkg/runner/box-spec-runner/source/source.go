package source

import (
	"bytes"
	"context"
	"log/slog"
	"sync"

	"github.com/dboxed/dboxed/pkg/boxspec"
	"sigs.k8s.io/yaml"
)

type BoxSpecSource struct {
	curSpecRaw []byte
	curSpec    *boxspec.BoxFile
	stopChan   chan struct{}
	Chan       chan *boxspec.BoxFile

	m sync.Mutex
}

func (h *BoxSpecSource) GetCurSpec() *boxspec.BoxFile {
	h.m.Lock()
	defer h.m.Unlock()
	return h.curSpec
}

func (h *BoxSpecSource) trySetNewSpec(ctx context.Context, raw []byte, initial bool, logError bool) error {
	if bytes.Equal(h.curSpecRaw, raw) {
		return nil
	}

	var spec *boxspec.BoxFile
	err := yaml.Unmarshal(raw, &spec)
	if err != nil {
		if logError {
			slog.WarnContext(ctx, "unmarshalling box spec failed", slog.Any("error", err))
		}
		return err
	}

	h.m.Lock()
	h.curSpecRaw = raw
	h.curSpec = spec
	h.m.Unlock()

	if !initial {
		h.Chan <- spec
	}
	return nil
}

func (h *BoxSpecSource) Cancel() {
	close(h.stopChan)
}
