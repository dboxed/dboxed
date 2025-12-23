package restic

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	securejoin "github.com/cyphar/filepath-securejoin"
)

type InitOpts struct {
}

func RunInit(ctx context.Context, env []string, opts InitOpts) error {
	args := []string{
		"init",
	}

	_, err := RunResticCommand(ctx, env, false, args)
	if err != nil {
		return err
	}
	return nil
}

type BackupOpts struct {
	Init      bool
	Host      *string
	WithAtime bool
	NoScan    bool
	Tags      []string
	Exclude   []string
}

func RunBackup(ctx context.Context, env []string, dir string, opts BackupOpts) (*Snapshot, error) {
	args := []string{
		"backup",
		"--one-file-system",
		"--json",
	}
	if opts.Init {
		args = append(args, "--init")
	}
	if opts.Host != nil {
		args = append(args, "--host", *opts.Host)
	}
	if opts.WithAtime {
		args = append(args, "--with-atime")
	}
	if opts.NoScan {
		args = append(args, "--no-scan")
	}
	for _, tag := range opts.Tags {
		args = append(args, "--tag", tag)
	}
	for _, exclude := range opts.Exclude {
		args = append(args, "--exclude", exclude)
	}

	args = append(args, dir)

	jsh := buildResticJsonStatusHandler(ctx)

	err := RunResticCommandStdoutCallback(ctx, env, args, jsh.handler)
	jsh.stop()
	if err != nil {
		return nil, err
	}

	summary := jsh.curSummary.Load()
	if summary == nil {
		return nil, fmt.Errorf("restic backup did not print summary")
	}
	snapshotIdI, ok := (*summary)["snapshot_id"]
	if !ok {
		return nil, fmt.Errorf("restic backup summary did not containt snapshot_id")
	}
	snapshotId, ok := snapshotIdI.(string)
	if !ok {
		return nil, fmt.Errorf("restic backup summary did not containt snapshot_id as a string")
	}

	snapshot, err := RunSingleSnapshots(ctx, env, SnapshotOpts{
		SnapshotIds: []string{snapshotId},
		NoLock:      true,
	})
	if err != nil {
		return nil, err
	}

	return snapshot, nil
}

type RestoreOpts struct {
	Delete bool
}

func RunRestore(ctx context.Context, env []string, snapshotId string, dir string, opts RestoreOpts) error {
	snapshot, err := RunSingleSnapshots(ctx, env, SnapshotOpts{
		SnapshotIds: []string{snapshotId},
		NoLock:      true,
	})
	if err != nil {
		return err
	}
	if len(snapshot.Paths) != 1 {
		return fmt.Errorf("unexpected number of snapshot paths")
	}

	tmpRestoreDir, err := os.MkdirTemp(dir, ".restore-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpRestoreDir)

	mvPath, err := securejoin.SecureJoin(tmpRestoreDir, snapshot.Paths[0])
	if err != nil {
		return err
	}

	args := []string{
		"restore",
		"--json",
	}
	if opts.Delete {
		args = append(args, "--delete")
	}

	args = append(args, snapshotId)
	args = append(args, "--target", tmpRestoreDir)

	jsh := buildResticJsonStatusHandler(ctx)
	err = RunResticCommandStdoutCallback(ctx, env, args, jsh.handler)
	jsh.stop()
	if err != nil {
		return err
	}

	des, err := os.ReadDir(mvPath)
	if err != nil {
		return err
	}
	for _, de := range des {
		// TODO remove this after some time
		if de.Name() == "lost+found" {
			continue
		}
		err = os.Rename(filepath.Join(mvPath, de.Name()), filepath.Join(dir, de.Name()))
		if err != nil {
			return fmt.Errorf("failed to move restored file/dir: %w", err)
		}
	}

	return nil
}

type SnapshotOpts struct {
	SnapshotIds []string
	NoCache     bool
	NoLock      bool
}

func RunSnapshots(ctx context.Context, env []string, opts SnapshotOpts) ([]Snapshot, error) {
	args := []string{
		"snapshots",
		"--json",
	}
	if opts.NoCache {
		args = append(args, "--no-cache")
	}
	if opts.NoLock {
		args = append(args, "--no-lock")
	}
	for _, id := range opts.SnapshotIds {
		args = append(args, id)
	}

	ret, err := RunResticCommandJson[[]Snapshot](ctx, env, args)
	if err != nil {
		return nil, err
	}
	return *ret, nil
}

func RunSingleSnapshots(ctx context.Context, env []string, opts SnapshotOpts) (*Snapshot, error) {
	snapshots, err := RunSnapshots(ctx, env, opts)
	if err != nil {
		return nil, err
	}
	if len(snapshots) != 1 {
		return nil, fmt.Errorf("unexpected number of snapshots returned: %d", len(snapshots))
	}
	snapshot := (snapshots)[0]
	return &snapshot, nil
}

type ForgetOpts struct {
	SnapshotIds []string
}

func RunForget(ctx context.Context, env []string, opts ForgetOpts) error {
	args := []string{
		"forget",
	}
	args = append(args, opts.SnapshotIds...)

	_, err := RunResticCommand(ctx, env, false, args)
	if err != nil {
		return err
	}
	return nil
}
