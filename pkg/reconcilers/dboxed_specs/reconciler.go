package dboxed_specs

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"path/filepath"
	"sync"
	"time"

	"github.com/dboxed/dboxed/pkg/reconcilers/base"
	"github.com/dboxed/dboxed/pkg/server/db/dmodel"
	"github.com/dboxed/dboxed/pkg/server/db/querier"
	"github.com/dboxed/dboxed/pkg/server/models/dboxed_specs"
	"github.com/dboxed/dboxed/pkg/server/resources/boxes_utils"
	"github.com/dboxed/dboxed/pkg/server/resources/volumes"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/kluctl/kluctl/lib/git/types"
	"sigs.k8s.io/yaml"
)

type reconciler struct {
}

type globalState struct {
	sshPools sync.Map
}

func NewDboxedSpecsReconciler() *base.Reconciler[*dmodel.DboxedSpec] {
	return base.NewReconciler(base.Config[*dmodel.DboxedSpec]{
		ReconcilerName:        "dboxed-specs",
		Reconciler:            &reconciler{},
		FullReconcileInterval: 15 * time.Second,
		NewGlobalState: func(ctx context.Context) any {
			return &globalState{}
		},
	})
}

func (r *reconciler) GetItem(ctx context.Context, id string) (*dmodel.DboxedSpec, error) {
	return dmodel.GetDboxedSpecById(querier.GetQuerier(ctx), nil, id, false)
}

func (r *reconciler) Reconcile(ctx context.Context, gs *dmodel.DboxedSpec, log *slog.Logger) base.ReconcileResult {
	u, err := types.ParseGitUrl(gs.GitUrl)
	if err != nil {
		return base.InternalError(err)
	}

	log = log.With(
		"repoKey", u.RepoKey().String(),
		"subdir", gs.Subdir,
		"specFile", gs.SpecFile,
	)

	mr, err := r.buildMirroredGitRepo(ctx, gs, log)
	if err != nil {
		return base.ErrorWithMessage(err, "failed to build mirrored git repo object")
	}
	err = mr.Lock()
	if err != nil {
		return base.InternalError(err)
	}
	defer mr.Unlock()

	if gs.DeletedAt.Valid {
		err = mr.Delete()
		if err != nil {
			slog.Error("failed to delete git mirror dir", "error", err)
		}
		return base.ReconcileResult{}
	}

	gt, result := r.openGitTree(gs, mr)
	if result.ExitReconcile() {
		return result
	}

	specsPath := filepath.Join(gs.Subdir, gs.SpecFile)
	specsBytes, err := r.loadFile(gt, specsPath)
	if err != nil {
		return base.ErrorWithMessage(err, "failed to load specs file %s", specsPath)
	}

	var specs dboxed_specs.DboxedSpecs
	err = yaml.Unmarshal(specsBytes, &specs)
	if err != nil {
		return base.ErrorWithMessage(err, "failed to unmarshal specs file %s. %s", specsPath, err.Error())
	}

	result = base.Transaction(ctx, func(ctx context.Context) base.ReconcileResult {
		return r.reconcileDboxedSpecs(ctx, gs, gt, &specs, log)
	})
	if result.ExitReconcile() {
		return result
	}

	return base.ReconcileResult{}
}

type typeAndName struct {
	t string
	n string
}

func (r *reconciler) reconcileDboxedSpecs(ctx context.Context, gs *dmodel.DboxedSpec, gt *object.Tree, specs *dboxed_specs.DboxedSpecs, log *slog.Logger) base.ReconcileResult {
	q := querier.GetQuerier(ctx)

	existingMappings, err := dmodel.ListDboxedSpecMappingForSpec(q, gs.WorkspaceID, gs.ID)
	if err != nil {
		return base.InternalError(err)
	}
	existingMappingsById := map[string]*dmodel.DboxedSpecMapping{}
	existingMappingsByTypeAndName := map[typeAndName]*dmodel.DboxedSpecMapping{}
	foundMappings := map[typeAndName]struct{}{}

	for _, m := range existingMappings {
		existingMappingsById[m.ObjectId] = &m
		existingMappingsByTypeAndName[typeAndName{t: m.ObjectType, n: m.ObjectName}] = &m
	}
	for name := range specs.Volumes {
		foundMappings[typeAndName{t: "volume", n: name}] = struct{}{}
	}
	for name := range specs.Boxes {
		foundMappings[typeAndName{t: "box", n: name}] = struct{}{}
	}

	doExitEarly := false
	for _, m := range existingMappings {
		if _, ok := foundMappings[typeAndName{t: m.ObjectType, n: m.ObjectName}]; ok {
			continue
		}
		deleted, result := r.deleteObject(ctx, &m, log)
		if result.ExitReconcile() {
			return result
		}
		if !deleted {
			doExitEarly = true
		}
	}
	if doExitEarly {
		return base.ReconcileResult{
			// let soft deletes finish first (they can only finish when the TX gets committed, so we return and let it commit)
			Requeue: true,
		}
	}

	for name, volume := range specs.Volumes {
		e := existingMappingsByTypeAndName[typeAndName{t: "volume", n: name}]
		result := r.reconcileSpecVolume(ctx, gs, name, &volume, e, log)
		if result.ExitReconcile() {
			return result
		}
	}
	for name, box := range specs.Boxes {
		e := existingMappingsByTypeAndName[typeAndName{t: "box", n: name}]
		result := r.reconcileSpecBox(ctx, gs, gt, name, &box, e, log)
		if result.ExitReconcile() {
			return result
		}
	}

	return base.ReconcileResult{}
}

func (r *reconciler) deleteObject(ctx context.Context, e *dmodel.DboxedSpecMapping, log *slog.Logger) (bool, base.ReconcileResult) {
	q := querier.GetQuerier(ctx)

	log = log.With("objectType", e.ObjectType, "objectName", e.ObjectName, "objectId", e.ObjectId)

	var getObjectFunc func() (dmodel.IsSoftDelete, error)
	var deleteObjectFunc func() error

	switch e.ObjectType {
	case "volume":
		getObjectFunc = func() (dmodel.IsSoftDelete, error) {
			return dmodel.GetVolumeById(q, &e.WorkspaceID, e.ObjectId, false)
		}
		deleteObjectFunc = func() error {
			return volumes.DeleteVolume(ctx, e.WorkspaceID, e.ObjectId)
		}
	case "box":
		getObjectFunc = func() (dmodel.IsSoftDelete, error) {
			return dmodel.GetBoxById(q, &e.WorkspaceID, e.ObjectId, false)
		}
		deleteObjectFunc = func() error {
			return boxes_utils.DeleteBox(ctx, e.WorkspaceID, e.ObjectId)
		}
	default:
		return false, base.InternalError(fmt.Errorf("unknown object type %s", e.ObjectType))
	}

	v, err := getObjectFunc()
	if err != nil {
		if querier.IsSqlNotFoundError(err) {
			log.InfoContext(ctx, "deleting dboxed spec mapping")
			err = querier.DeleteOneByStruct(q, e)
			if err != nil {
				return false, base.InternalError(err)
			}
			return true, base.ReconcileResult{}
		}
		return false, base.InternalError(err)
	}
	if v.GetDeletedAt() != nil {
		return false, base.ReconcileResult{}
	}

	log.InfoContext(ctx, "deleting object")
	err = deleteObjectFunc()
	if err != nil {
		return false, base.ErrorWithMessage(err, "could not delete %s with id %s", e.ObjectType, e.ObjectId)
	}

	return false, base.ReconcileResult{}
}

func (r *reconciler) createMapping(ctx context.Context, workspaceId string, spec *dmodel.DboxedSpec, recreateKey string, objectType string, objectId string, objectName string, specFragment any) (*dmodel.DboxedSpecMapping, base.ReconcileResult) {
	q := querier.GetQuerier(ctx)

	specStr, err := json.Marshal(specFragment)
	if err != nil {
		return nil, base.InternalError(err)
	}

	m := &dmodel.DboxedSpecMapping{
		OwnedByWorkspace: dmodel.OwnedByWorkspace{
			WorkspaceID: workspaceId,
		},
		SpecId:       spec.ID,
		ObjectType:   objectType,
		ObjectId:     objectId,
		ObjectName:   objectName,
		RecreateKey:  recreateKey,
		SpecFragment: string(specStr),
	}
	err = m.Create(q)
	if err != nil {
		return nil, base.ErrorWithMessage(err, "failed to create dboxed spec mapping")
	}
	return m, base.ReconcileResult{}
}
