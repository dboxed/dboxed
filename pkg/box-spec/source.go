package box_spec

import (
	"bytes"
	"context"
	"log/slog"
	"sync"

	"github.com/dboxed/dboxed/pkg/types"
	"sigs.k8s.io/yaml"
)

type BoxSpecSource struct {
	curSpecRaw []byte
	curSpec    *types.BoxFile
	stopChan   chan struct{}
	Chan       chan *types.BoxFile

	m sync.Mutex
}

func (h *BoxSpecSource) GetCurSpec() *types.BoxFile {
	h.m.Lock()
	defer h.m.Unlock()
	return h.curSpec
}

func (h *BoxSpecSource) trySetNewSpec(ctx context.Context, raw []byte, initial bool, logError bool) error {
	if bytes.Equal(h.curSpecRaw, raw) {
		return nil
	}

	var spec types.BoxFile
	err := yaml.Unmarshal(raw, &spec)
	if err != nil {
		if logError {
			slog.WarnContext(ctx, "unmarshalling box spec failed", slog.Any("error", err))
		}
		return err
	}

	h.m.Lock()
	defer h.m.Unlock()
	h.curSpecRaw = raw
	h.curSpec = &spec
	if !initial {
		h.Chan <- h.curSpec
	}
	return nil
}

func (h *BoxSpecSource) Cancel() {
	close(h.stopChan)
}
