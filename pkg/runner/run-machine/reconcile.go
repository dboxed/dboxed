package run_machine

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/dboxed/dboxed/pkg/clients"
	run_sandbox "github.com/dboxed/dboxed/pkg/runner/run-sandbox"
	"github.com/dboxed/dboxed/pkg/runner/sandbox"
	"github.com/dboxed/dboxed/pkg/server/models"
	"github.com/dboxed/dboxed/pkg/util/command_helper"
	"github.com/opencontainers/runc/libcontainer"
)

func (rn *RunMachine) reconcileMachine(ctx context.Context, boxes []models.Box) error {
	slog.DebugContext(ctx, "starting reconcile of machine")

	didSetMachineStatusReconciling := false
	doSetMachineStatusReconciling := func() {
		if didSetMachineStatusReconciling {
			return
		}
		rn.updateMachineStatusSimple(ctx, "reconciling", true)
		didSetMachineStatusReconciling = true
	}

	sandboxBaseDir := run_sandbox.GetSandboxDir(rn.WorkDir, "")
	sandboxInfos, err := sandbox.ListSandboxes(sandboxBaseDir)
	if err != nil {
		return err
	}
	sandboxesByBoxId := map[string]*sandbox.SandboxInfo{}
	sandboxStatusById := map[string]libcontainer.Status{}
	boxesById := map[string]*models.Box{}
	for _, si := range sandboxInfos {
		sandboxesByBoxId[si.Box.ID] = &si
		s := sandbox.Sandbox{
			Debug:       rn.Debug,
			HostWorkDir: rn.WorkDir,
			SandboxDir:  filepath.Join(sandboxBaseDir, si.Box.ID),
		}
		cs, err := s.GetSandboxContainerStatus()
		if err != nil {
			return fmt.Errorf("failed to determine sandbox container status: %w", err)
		}
		sandboxStatusById[si.Box.ID] = cs
	}
	for _, box := range boxes {
		boxesById[box.ID] = &box
	}

	for _, box := range boxes {
		log := slog.With("boxId", box.ID, "boxName", box.Name)

		if !box.Enabled {
			continue
		}

		si, ok := sandboxesByBoxId[box.ID]
		if ok {
			log = log.With("sandboxId", si.SandboxId)
		}

		if !ok {
			log.InfoContext(ctx, "starting sandbox for new box")
		} else {
			cs := sandboxStatusById[box.ID]
			if cs != libcontainer.Running {
				log.InfoContext(ctx, "sandbox is not in running state, restarting", "state", cs.String())
				ok = false
			}
		}

		if !ok {
			doSetMachineStatusReconciling()
			err = rn.startSandbox(ctx, &box)
			if err != nil {
				return err
			}
		}
	}

	for _, si := range sandboxInfos {
		log := slog.With("boxId", si.Box.ID, "boxName", si.Box.Name, "sandboxId", si.SandboxId)

		_, ok := boxesById[si.Box.ID]
		if !ok {
			doSetMachineStatusReconciling()
			log.InfoContext(ctx, "box removed from machine, stopping and removing sandbox")

			err = rn.stopSandbox(ctx, si)
			if err != nil {
				return err
			}
			err = rn.removeSandbox(ctx, si)
			if err != nil {
				return err
			}
		}
	}

	rn.updateMachineStatusSimple(ctx, "running", true)
	return nil
}

func (rn *RunMachine) startSandbox(ctx context.Context, box *models.Box) error {
	mc := clients.MachineClient{Client: rn.Client}

	token, err := mc.CreateBoxToken(ctx, rn.MachineId, box.ID)
	if err != nil {
		return err
	}

	selfExe, err := os.Executable()
	if err != nil {
		return err
	}

	args := []string{
		"sandbox",
		"run",
		box.ID,
		"--work-dir", rn.WorkDir,
		"--infra-image", rn.InfraImage,
		"--veth-cidr", rn.VethCidr,
	}
	if rn.Debug {
		args = append(args, "--debug")
	}

	env := os.Environ()
	env = append(env, fmt.Sprintf("DBOXED_API_URL=%s", rn.Client.GetClientAuth(true).ApiUrl))
	env = append(env, fmt.Sprintf("DBOXED_API_TOKEN=%s", *token.Token))

	cmd := command_helper.CommandHelper{
		Command: selfExe,
		Args:    args,
		Env:     env,
		LogCmd:  true,
	}

	err = cmd.Run(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (rn *RunMachine) stopSandbox(ctx context.Context, si sandbox.SandboxInfo) error {
	selfExe, err := os.Executable()
	if err != nil {
		return err
	}

	args := []string{
		"sandbox",
		"stop",
		si.Box.ID,
		"--work-dir", rn.WorkDir,
	}
	if rn.Debug {
		args = append(args, "--debug")
	}

	cmd := command_helper.CommandHelper{
		Command: selfExe,
		Args:    args,
		LogCmd:  true,
	}

	err = cmd.Run(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (rn *RunMachine) removeSandbox(ctx context.Context, si sandbox.SandboxInfo) error {
	selfExe, err := os.Executable()
	if err != nil {
		return err
	}

	args := []string{
		"sandbox",
		"rm",
		si.Box.ID,
		"--work-dir", rn.WorkDir,
	}
	if rn.Debug {
		args = append(args, "--debug")
	}

	cmd := command_helper.CommandHelper{
		Command: selfExe,
		Args:    args,
		LogCmd:  true,
	}
	err = cmd.Run(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (rn *RunMachine) shutdown(ctx context.Context) error {
	return rn.reconcileMachine(ctx, nil)
}
