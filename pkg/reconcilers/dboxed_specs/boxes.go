package dboxed_specs

import (
	"context"
	"encoding/json"
	"log/slog"
	"path/filepath"
	"sort"

	"github.com/dboxed/dboxed/pkg/reconcilers/base"
	"github.com/dboxed/dboxed/pkg/server/db/dmodel"
	"github.com/dboxed/dboxed/pkg/server/db/querier"
	"github.com/dboxed/dboxed/pkg/server/models"
	"github.com/dboxed/dboxed/pkg/server/models/dboxed_specs"
	"github.com/dboxed/dboxed/pkg/server/resources/boxes_utils"
	"github.com/dboxed/dboxed/pkg/util"
	"github.com/go-git/go-git/v5/plumbing/object"
)

func (r *reconciler) reconcileSpecBox(ctx context.Context, gs *dmodel.DboxedSpec, gt *object.Tree, name string, box *dboxed_specs.Box, e *dmodel.DboxedSpecMapping, log *slog.Logger) base.ReconcileResult {
	q := querier.GetQuerier(ctx)

	log = log.With("boxName", name)

	if e == nil {
		newE, result := r.createBox(ctx, gs, name, box, log)
		if result.ExitReconcile() {
			return result
		}
		e = newE
	}

	var oldFragment *dboxed_specs.OnlyRecreate
	err := json.Unmarshal([]byte(e.SpecFragment), &oldFragment)
	if err != nil {
		return base.InternalError(err)
	}
	if box.Recreate != oldFragment.Recreate {
		deleted, result := r.deleteObject(ctx, e, log)
		if result.ExitReconcile() {
			return result
		}
		if deleted {
			newE, result := r.createBox(ctx, gs, name, box, log)
			if result.ExitReconcile() {
				return result
			}
			e = newE
		} else {
			return base.ReconcileResult{}
		}
	}

	dbBox, err := dmodel.GetBoxById(q, &gs.WorkspaceID, e.ObjectId, true)
	if err != nil {
		return base.InternalError(err)
	}
	log = log.With("boxId", dbBox.ID)

	result := r.reconcileBoxVolumeAttachments(ctx, gs, box, dbBox, log)
	if result.ExitReconcile() {
		return result
	}
	result = r.reconcileBoxVolumeComposeProjects(ctx, gs, gt, box, dbBox, log)
	if result.ExitReconcile() {
		return result
	}
	result = r.reconcileBoxLoadBalancerServices(ctx, gs, box, dbBox, log)
	if result.ExitReconcile() {
		return result
	}
	result = r.reconcileBoxMachine(ctx, gs, box, dbBox, log)
	if result.ExitReconcile() {
		return result
	}

	return base.ReconcileResult{}
}

func (r *reconciler) createBox(ctx context.Context, gs *dmodel.DboxedSpec, name string, box *dboxed_specs.Box, log *slog.Logger) (*dmodel.DboxedSpecMapping, base.ReconcileResult) {
	q := querier.GetQuerier(ctx)

	var err error
	var network *dmodel.Network
	if box.Network != nil {
		network, err = dmodel.GetNetworkByName(q, gs.WorkspaceID, *box.Network, true)
		if err != nil {
			return nil, base.ErrorWithMessage(err, "failed to retrieve network with name '%s'", *box.Network)
		}
	}

	createArgs := models.CreateBox{
		Name: name,
	}
	if network != nil {
		createArgs.Network = &network.ID
	}

	log.InfoContext(ctx, "creating box")
	dbBox, err := boxes_utils.CreateBox(ctx, gs.WorkspaceID, createArgs, dmodel.BoxTypeDboxedSpec)
	if err != nil {
		return nil, base.InternalError(err)
	}

	e, result := r.createMapping(ctx, gs.WorkspaceID, gs, box.Recreate, "box", dbBox.ID, name, box)
	if result.ExitReconcile() {
		return nil, result
	}
	return e, base.ReconcileResult{}
}

func (r *reconciler) reconcileBoxVolumeAttachments(ctx context.Context, gs *dmodel.DboxedSpec, box *dboxed_specs.Box, dbBox *dmodel.Box, log *slog.Logger) base.ReconcileResult {
	q := querier.GetQuerier(ctx)

	existingAttachments, err := dmodel.ListBoxVolumeAttachments(q, dbBox.ID)
	if err != nil {
		return base.InternalError(err)
	}
	existingAttachmentsMap := map[string]*dmodel.BoxVolumeAttachmentWithJoins{}
	for _, ba := range existingAttachments {
		existingAttachmentsMap[ba.Volume.Name] = &ba
	}

	foundAttachments := map[string]struct{}{}
	for _, a := range box.VolumeAttachments {
		foundAttachments[a.Volume] = struct{}{}

		log := log.With("volumeName", a.Volume)

		err = boxes_utils.CheckVolumeAttachmentParams(a.RootGid, a.RootGid, a.RootMode)
		if err != nil {
			return base.InternalError(err)
		}

		ba := existingAttachmentsMap[a.Volume]
		if ba == nil {
			volume, err := dmodel.GetVolumeByName(q, gs.WorkspaceID, a.Volume, true)
			if err != nil {
				return base.ErrorWithMessage(err, "failed to get volume %s", a.Volume)
			}

			log.InfoContext(ctx, "attaching volume to box")
			err = boxes_utils.AttachVolume(ctx, dbBox, models.AttachVolumeRequest{
				VolumeId: volume.ID,
				RootUid:  a.RootUid,
				RootGid:  a.RootGid,
				RootMode: a.RootMode,
			})
			if err != nil {
				return base.ErrorWithMessage(err, "failed to attach volume %s", a.Volume)
			}
		} else {
			err = ba.Update(q, a.RootUid, a.RootGid, a.RootMode)
			if err != nil {
				return base.ErrorWithMessage(err, "failed to update volume attachment %s", a.Volume)
			}
		}
	}

	for _, ba := range existingAttachmentsMap {
		log := log.With("volumeName", ba.Volume.Name, "volumeId", ba.Volume.ID)
		if _, ok := foundAttachments[ba.Volume.Name]; ok {
			continue
		}
		log.InfoContext(ctx, "detaching volume from box")
		err = boxes_utils.DetachVolume(ctx, dbBox, ba.Volume.ID)
		if err != nil {
			return base.ErrorWithMessage(err, "failed to detach volume %s from box", ba.Volume.Name)
		}
	}

	return base.ReconcileResult{}
}

func (r *reconciler) reconcileBoxVolumeComposeProjects(ctx context.Context, gs *dmodel.DboxedSpec, gt *object.Tree, box *dboxed_specs.Box, dbBox *dmodel.Box, log *slog.Logger) base.ReconcileResult {
	q := querier.GetQuerier(ctx)

	existingComposeProjects, err := dmodel.ListBoxComposeProjects(q, dbBox.ID)
	if err != nil {
		return base.InternalError(err)
	}
	existingComposeProjectsMap := map[string]*dmodel.BoxComposeProject{}
	for _, ba := range existingComposeProjects {
		existingComposeProjectsMap[ba.Name] = &ba
	}

	foundComposeProjects := map[string]struct{}{}
	for name, cp := range box.ComposeProjects {
		foundComposeProjects[name] = struct{}{}

		log := log.With("composeName", name)

		path := filepath.Join(gs.Subdir, cp.File)
		cpContentBytes, err := r.loadFile(gt, path)
		if err != nil {
			return base.ErrorWithMessage(err, "failed load compose file %s", path)
		}
		cpContent := string(cpContentBytes)

		ecp := existingComposeProjectsMap[name]
		if ecp == nil {
			log.InfoContext(ctx, "adding compose project")
			err = boxes_utils.CreateComposeProject(ctx, dbBox, models.CreateBoxComposeProject{
				Name:           name,
				ComposeProject: cpContent,
			})
			if err != nil {
				return base.ErrorWithMessage(err, "failed to add compose project %s", name)
			}
		} else {
			if cpContent != ecp.ComposeProject {
				err = boxes_utils.UpdateComposeProject(ctx, dbBox, name, cpContent)
				if err != nil {
					return base.ErrorWithMessage(err, "failed to update compose project %s", name)
				}
			}
		}
	}

	for _, ecp := range existingComposeProjectsMap {
		log := log.With("composeName", ecp.Name)
		if _, ok := foundComposeProjects[ecp.Name]; ok {
			continue
		}
		log.InfoContext(ctx, "removing compose project from box")
		err = boxes_utils.DeleteComposeProject(ctx, dbBox, ecp.Name)
		if err != nil {
			return base.ErrorWithMessage(err, "failed to delete compose project %s from box", ecp.Name)
		}
	}

	return base.ReconcileResult{}
}

func (r *reconciler) reconcileBoxLoadBalancerServices(ctx context.Context, gs *dmodel.DboxedSpec, box *dboxed_specs.Box, dbBox *dmodel.Box, log *slog.Logger) base.ReconcileResult {
	q := querier.GetQuerier(ctx)

	existingLBServices, err := dmodel.ListLoadBalancerServices(q, dbBox.ID)
	if err != nil {
		return base.InternalError(err)
	}

	loadBalancerById := map[string]*dmodel.LoadBalancer{}
	loadBalancerByName := map[string]*dmodel.LoadBalancer{}

	getLoadBalancerByName := func(name string) (*dmodel.LoadBalancer, error) {
		if lb, ok := loadBalancerByName[name]; ok {
			return lb, nil
		}
		lb, err := dmodel.GetLoadBalancerByName(q, &gs.WorkspaceID, name, true)
		if err != nil {
			return nil, err
		}
		loadBalancerByName[lb.Name] = lb
		loadBalancerById[lb.ID] = lb
		return lb, nil
	}
	getLoadBalancerById := func(id string) (*dmodel.LoadBalancer, error) {
		if lb, ok := loadBalancerById[id]; ok {
			return lb, nil
		}
		lb, err := dmodel.GetLoadBalancerById(q, &gs.WorkspaceID, id, true)
		if err != nil {
			return nil, err
		}
		loadBalancerByName[lb.Name] = lb
		loadBalancerById[lb.ID] = lb
		return lb, nil
	}

	var oldLBServices []dboxed_specs.LoadBalancerService
	for _, lbs := range existingLBServices {
		lb, err := getLoadBalancerById(lbs.LoadBalancerId)
		if err != nil {
			return base.ErrorWithMessage(err, "failed to get load balancer with id %s", lbs.LoadBalancerId)
		}
		x := dboxed_specs.LoadBalancerService{
			LoadBalancer: lb.Name,
			Host:         lbs.Hostname,
			PathPrefix:   lbs.PathPrefix,
			Port:         lbs.Port,
			Description:  lbs.Description,
		}
		oldLBServices = append(oldLBServices, x)
	}
	sort.Slice(oldLBServices, func(i, j int) bool {
		return oldLBServices[i].LoadBalancer < oldLBServices[j].LoadBalancer
	})
	sort.Slice(box.LoadBalancerServices, func(i, j int) bool {
		return box.LoadBalancerServices[i].LoadBalancer < box.LoadBalancerServices[j].LoadBalancer
	})

	if util.EqualsViaJson(oldLBServices, box.LoadBalancerServices) {
		return base.ReconcileResult{}
	}

	for _, lbs := range existingLBServices {
		err = querier.DeleteOneByStruct(q, lbs)
		if err != nil {
			return base.InternalError(err)
		}
	}
	for _, lbs := range box.LoadBalancerServices {
		lb, err := getLoadBalancerByName(lbs.LoadBalancer)
		if err != nil {
			return base.ErrorWithMessage(err, "did not find load balancer with name %s", lbs.LoadBalancer)
		}
		dbLbs := dmodel.LoadBalancerService{
			LoadBalancerId: lb.ID,
			BoxID:          dbBox.ID,
			Description:    lbs.Description,
			Hostname:       lbs.Host,
			PathPrefix:     lbs.PathPrefix,
			Port:           lbs.Port,
		}
		err = dbLbs.Create(q)
		if err != nil {
			return base.InternalError(err)
		}
	}

	return base.ReconcileResult{}
}

func (r *reconciler) reconcileBoxMachine(ctx context.Context, gs *dmodel.DboxedSpec, box *dboxed_specs.Box, dbBox *dmodel.Box, log *slog.Logger) base.ReconcileResult {
	q := querier.GetQuerier(ctx)

	var oldMachineId, newMachineId *string

	if dbBox.MachineID != nil {
		oldMachineId = dbBox.MachineID
	}
	if box.Machine != nil {
		machine, err := dmodel.GetMachineByName(q, dbBox.WorkspaceID, *box.Machine, true)
		if err != nil {
			return base.ErrorWithMessage(err, "failed to get machine with name '%s'", *box.Machine)
		}
		newMachineId = &machine.ID
	}

	if util.PtrEquals(oldMachineId, newMachineId) {
		return base.ReconcileResult{}
	}

	if !dbBox.MachineFromSpec && oldMachineId != nil {
		if newMachineId == nil {
			return base.ReconcileResult{}
		}
		return base.ErrorFromMessage("box was already manually attached to a machine, can't override this via dboxed specs")
	}

	log.InfoContext(ctx, "updating machine for spec box", "newMachineId", newMachineId)

	fromSpec := newMachineId != nil
	err := dbBox.UpdateMachineID(q, newMachineId, fromSpec)
	if err != nil {
		return base.InternalError(err)
	}

	return base.ReconcileResult{}
}
