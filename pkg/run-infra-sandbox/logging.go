package run_infra_sandbox

import (
	"context"
	"github.com/koobox/unboxed/pkg/types"
	"path/filepath"
)

func (rn *RunInfraSandbox) initLogsPublishing(ctx context.Context) error {
	// the tails db used here is the same as the one used in start-box, so we will actually
	// continue where start-box stopped
	err := rn.logsPublisher.Start(ctx, rn.conf.BoxSpec, filepath.Join(types.LogsDir, types.LogsTailDbFilename))
	if err != nil {
		return err
	}

	err = rn.logsPublisher.PublishUnboxedLogsDir("/var/lib/unboxed/logs")
	if err != nil {
		return err
	}
	err = rn.logsPublisher.PublishDockerLogsDir("/var/lib/docker")
	if err != nil {
		return err
	}

	return nil
}
