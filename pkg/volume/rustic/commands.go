package rustic

import "context"

type InitOpts struct {
}

func RunInit(ctx context.Context, config RusticConfig, opts InitOpts) error {
	args := []string{
		"init",
	}

	_, err := RunRusticCommand(ctx, config, false, args)
	if err != nil {
		return err
	}
	return nil
}

type BackupOpts struct {
	Init      bool
	Host      *string
	AsPath    *string
	WithAtime bool
	NoScan    bool
	Tags      []string
}

func RunBackup(ctx context.Context, config RusticConfig, dir string, opts BackupOpts) (*Snapshot, error) {
	args := []string{
		"backup",
		"--json",
	}
	if opts.Init {
		args = append(args, "--init")
	}
	if opts.Host != nil {
		args = append(args, "--host", *opts.Host)
	}
	if opts.AsPath != nil {
		args = append(args, "--as-path", *opts.AsPath)
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

	args = append(args, dir)

	rs, err := RunRusticCommandJson[Snapshot](ctx, config, args)
	if err != nil {
		return nil, err
	}
	return rs, nil
}

type RestoreOpts struct {
	NumericId bool
	Delete    bool
}

func RunRestore(ctx context.Context, config RusticConfig, snapshotId string, dir string, opts RestoreOpts) error {
	args := []string{
		"restore",
	}
	if opts.NumericId {
		args = append(args, "--numeric-id")
	}
	if opts.Delete {
		args = append(args, "--delete")
	}

	args = append(args, snapshotId)
	args = append(args, dir)

	_, err := RunRusticCommand(ctx, config, false, args)
	if err != nil {
		return err
	}
	return err
}

type SnapshotOpts struct {
	SnapshotIds []string
	NoCache     bool
}

func RunSnapshots(ctx context.Context, config RusticConfig, opts SnapshotOpts) ([]Snapshot, error) {
	args := []string{
		"snapshots",
		"--json",
		"--no-progress",
		"--group-by", "",
	}
	if opts.NoCache {
		args = append(args, "--no-cache")
	}
	for _, id := range opts.SnapshotIds {
		args = append(args, id)
	}

	ret, err := RunRusticCommandJson[[]Snapshot](ctx, config, args)
	if err != nil {
		return nil, err
	}
	return *ret, nil
}

type ForgetOpts struct {
	SnapshotIds []string
}

func RunForget(ctx context.Context, config RusticConfig, opts ForgetOpts) error {
	args := []string{
		"forget",
	}
	args = append(args, opts.SnapshotIds...)

	_, err := RunRusticCommand(ctx, config, false, args)
	if err != nil {
		return err
	}
	return nil
}
