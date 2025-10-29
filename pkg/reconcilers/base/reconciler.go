package base

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/dboxed/dboxed/pkg/server/config"
	"github.com/dboxed/dboxed/pkg/server/db/dmodel"
	querier2 "github.com/dboxed/dboxed/pkg/server/db/querier"
	"k8s.io/client-go/util/workqueue"
)

type ReconcileImpl[T dmodel.HasReconcileStatusAndSoftDelete] interface {
	GetItem(ctx context.Context, id int64) (T, error)

	Reconcile(ctx context.Context, v T, log *slog.Logger) ReconcileResult
}

type Config[T dmodel.HasReconcileStatusAndSoftDelete] struct {
	ServerConfig config.Config

	ReconcilerName string

	ChangeCheckInterval   time.Duration
	FullReconcileInterval time.Duration
	Parallel              int

	Impl ReconcileImpl[T]
}

type Reconciler[T dmodel.HasReconcileStatusAndSoftDelete] struct {
	config Config[T]

	tableName  string
	didInitial bool
	workQueue  *workqueue.Typed[workQueueItem]

	lastChangeId int64

	log *slog.Logger
}

type workQueueItem struct {
	id int64
}

func NewReconciler[T dmodel.HasReconcileStatusAndSoftDelete](config Config[T]) *Reconciler[T] {
	if config.ChangeCheckInterval == 0 {
		config.ChangeCheckInterval = time.Second * 1
	}
	if config.Parallel == 0 {
		config.Parallel = 4
	}

	r := &Reconciler[T]{
		config:       config,
		workQueue:    workqueue.NewTyped[workQueueItem](),
		lastChangeId: -1,
		log:          slog.With(slog.Any("reconciler", config.ReconcilerName)),
	}
	return r
}

func (r *Reconciler[T]) Run(ctx context.Context) error {
	r.tableName = querier2.GetTableName[T]()

	for range r.config.Parallel {
		go r.runQueue(ctx)
	}

	nextFullCheck := time.Now().Add(r.config.FullReconcileInterval)
	for {
		fullCheck := false
		if r.config.FullReconcileInterval != 0 && time.Now().After(nextFullCheck) {
			fullCheck = true
			nextFullCheck = time.Now().Add(r.config.FullReconcileInterval)
		}

		r.findChanges(ctx, fullCheck)

		select {
		case <-ctx.Done():
			r.workQueue.ShutDown()
			return ctx.Err()
		case <-time.After(r.config.ChangeCheckInterval):
		}
	}
}

func (r *Reconciler[T]) findChanges(ctx context.Context, fullCheck bool) {
	q := querier2.GetQuerier(ctx)

	toQueue := map[int64]workQueueItem{}
	if !r.didInitial || fullCheck {
		maxId, err := dmodel.GetMaxChangeTrackingId[T](q)
		if err != nil {
			if !querier2.IsSqlNotFoundError(err) {
				slog.ErrorContext(ctx, "error in GetMaxChangeTrackingId", slog.Any("error", err))
				return
			}
			maxId = -1
		}
		r.lastChangeId = maxId

		allIds, err := dmodel.GetAllIds[T](q)
		if err != nil {
			slog.ErrorContext(ctx, "error in GetAllIds", slog.Any("error", err))
			return
		}
		r.didInitial = true

		for _, id := range allIds {
			if _, ok := toQueue[id]; !ok {
				toQueue[id] = workQueueItem{id: id}
			}
		}
	} else {
		changedItems, err := dmodel.FindChanges[T](q, r.lastChangeId)
		if err != nil {
			slog.ErrorContext(ctx, "error in ListChangeTracking", slog.Any("error", err))
			return
		}
		for _, ci := range changedItems {
			r.lastChangeId = ci.ID
			toQueue[ci.EntityID] = workQueueItem{id: ci.EntityID}
		}
	}

	for _, wi := range toQueue {
		r.workQueue.Add(wi)
	}
}

func (r *Reconciler[T]) runQueue(ctx context.Context) {
	for {
		if !r.runQueueOnce(ctx) {
			return
		}
	}
}

func (r *Reconciler[T]) runQueueOnce(ctx context.Context) bool {
	q := querier2.GetQuerier(ctx)

	item, shutdown := r.workQueue.Get()
	if shutdown {
		return false
	}
	defer r.workQueue.Done(item)

	log := r.log.With(
		slog.Any("id", item.id),
	)

	v, err := r.config.Impl.GetItem(ctx, item.id)
	if err != nil {
		if !querier2.IsSqlNotFoundError(err) {
			log.ErrorContext(ctx, "error getting reconcile item", slog.Any("error", err))
		}
		return true
	}

	result := r.config.Impl.Reconcile(ctx, v, log)
	if result.Retry {
		r.workQueue.Add(item)
		return true
	}
	SetReconcileResult(ctx, log, v, result)

	if result.Error == nil && v.GetDeletedAt() != nil {
		if len(v.GetFinalizers()) == 0 {
			log.InfoContext(ctx, fmt.Sprintf("finally deleting %s", r.tableName))
			err = querier2.DeleteOneByStruct(q, v)
			if err != nil {
				SetReconcileResult(ctx, log, v, InternalError(err))
			}
		}
	}

	return true
}
