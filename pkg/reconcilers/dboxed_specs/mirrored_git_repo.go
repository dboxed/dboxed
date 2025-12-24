package dboxed_specs

import (
	"context"
	"io"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/dboxed/dboxed/pkg/reconcilers/base"
	"github.com/dboxed/dboxed/pkg/server/config"
	"github.com/dboxed/dboxed/pkg/server/db/dmodel"
	"github.com/dboxed/dboxed/pkg/server/db/querier"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/gobwas/glob"
	"github.com/kluctl/kluctl/lib/git"
	"github.com/kluctl/kluctl/lib/git/auth"
	"github.com/kluctl/kluctl/lib/git/messages"
	ssh_pool "github.com/kluctl/kluctl/lib/git/ssh-pool"
	"github.com/kluctl/kluctl/lib/git/types"
)

func (r *reconciler) buildMirroredGitRepo(ctx context.Context, gs *dmodel.DboxedSpec, log *slog.Logger) (*git.MirroredGitRepo, error) {
	cfg := config.GetConfig(ctx)
	g := base.GetGlobalState[globalState](ctx)

	gitUrl, err := types.ParseGitUrl(gs.GitUrl)
	if err != nil {
		return nil, err
	}

	baseDir := filepath.Join(cfg.GitMirrorDir, gs.WorkspaceID)
	err = os.MkdirAll(baseDir, 0700)
	if err != nil {
		return nil, err
	}

	sshPool1, _ := g.sshPools.LoadOrStore(gs.WorkspaceID, &ssh_pool.SshPool{})
	sshPool := sshPool1.(*ssh_pool.SshPool)

	authProviders, err := r.buildAuthProviders(ctx, gs, log)
	if err != nil {
		return nil, err
	}

	mr, err := git.NewMirroredGitRepo(ctx, *gitUrl, baseDir, sshPool, authProviders)
	if err != nil {
		return nil, err
	}
	return mr, nil
}

func (r *reconciler) buildAuthProviders(ctx context.Context, gs *dmodel.DboxedSpec, log *slog.Logger) (*auth.GitAuthProviders, error) {
	q := querier.GetQuerier(ctx)
	gitCreds, err := dmodel.ListGitCredentialsForWorkspace(q, gs.WorkspaceID)
	if err != nil {
		return nil, err
	}

	messageCallbacks := messages.MessageCallbacks{
		WarningFn: func(s string) { log.WarnContext(ctx, s) },
		TraceFn:   func(s string) { log.DebugContext(ctx, s) },
	}

	gitAuthList := auth.ListAuthProvider{
		MessageCallbacks: messageCallbacks,
	}
	for _, gc := range gitCreds {
		e := auth.AuthEntry{
			Host:             gc.Host,
			IgnoreKnownHosts: true,
		}
		if gc.PathGlob != "" {
			e.PathGlob, err = glob.Compile(gc.PathGlob, '/')
			if err != nil {
				return nil, err
			}
		}
		if gc.Username != nil {
			e.Username = *gc.Username
		}
		if gc.Password != nil {
			e.Password = *gc.Password
		}
		if gc.SshKey != nil {
			e.SshKey = []byte(*gc.SshKey)
		}
		gitAuthList.AddEntry(e)
	}

	var ret auth.GitAuthProviders
	ret.RegisterAuthProvider(&gitAuthList, false)
	return &ret, nil
}

func (r *reconciler) openGitTree(gs *dmodel.DboxedSpec, mr *git.MirroredGitRepo) (*object.Tree, base.ReconcileResult) {
	err := mr.Update()
	if err != nil {
		return nil, base.ErrorWithMessage(err, "failed to update mirrored git repo")
	}

	ref := gs.GetGitRef()
	if ref == nil {
		ref, err = mr.DefaultRef()
		if err != nil {
			return nil, base.ErrorWithMessage(err, "failed to determine default branch")
		}
	}

	refs, err := mr.RemoteRefHashesMap()
	if err != nil {
		return nil, base.ErrorWithMessage(err, "failed to list refs")
	}
	commit, err := git.FindCommitByRef(mr, refs, *ref)
	if err != nil {
		return nil, base.ErrorWithMessage(err, "failed to find commit for ref %s", ref.String())
	}

	gt, err := mr.GetGitTreeByCommit(commit)
	if err != nil {
		return nil, base.ErrorWithMessage(err, "failed to open git tree by commit %s", commit)
	}

	return gt, base.ReconcileResult{}
}

func (r *reconciler) loadFile(gt *object.Tree, path string) ([]byte, error) {
	f, err := gt.File(path)
	if err != nil {
		return nil, err
	}
	rdr, err := f.Reader()
	if err != nil {
		return nil, err
	}
	defer rdr.Close()
	return io.ReadAll(rdr)
}
