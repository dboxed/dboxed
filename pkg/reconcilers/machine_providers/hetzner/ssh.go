package hetzner

import (
	"context"
	"fmt"

	"github.com/dboxed/dboxed/pkg/reconcilers/base"
	"github.com/dboxed/dboxed/pkg/server/cloud_utils"
	"github.com/dboxed/dboxed/pkg/server/config"
	"github.com/hetznercloud/hcloud-go/v2/hcloud"
	"golang.org/x/crypto/ssh"
)

func (r *Reconciler) reconcileSshKey(ctx context.Context) base.ReconcileResult {
	config := config.GetConfig(ctx)
	if r.mp.SshKeyPublic == nil {
		r.sshKeyId = -1
		return base.ReconcileResult{}
	}
	keyName := fmt.Sprintf("%s--%s-%d", config.InstanceName, r.mp.Name, r.mp.ID)

	pk, _, _, _, err := ssh.ParseAuthorizedKey([]byte(*r.mp.SshKeyPublic))
	if err != nil {
		return base.ErrorWithMessage(err, "failed to parse SSH public key: %s", err.Error())
	}
	fingerprint := ssh.FingerprintLegacyMD5(pk)

	k, _, err := r.hcloudClient.SSHKey.GetByFingerprint(ctx, fingerprint)
	if err != nil {
		return base.ErrorWithMessage(err, "failed to get SSH key from Hetzner: %s", err.Error())
	}
	if k != nil {
		r.sshKeyId = k.ID
		return base.ReconcileResult{}
	}
	k, _, err = r.hcloudClient.SSHKey.Create(ctx, hcloud.SSHKeyCreateOpts{
		Name:      keyName,
		PublicKey: *r.mp.SshKeyPublic,
		Labels:    cloud_utils.BuildCloudBaseTags(r.mp.ID, r.mp.WorkspaceID),
	})
	if err != nil {
		return base.ErrorWithMessage(err, "failed to create SSH key on Hetzner: %s", err.Error())
	}
	r.sshKeyId = k.ID

	return base.ReconcileResult{}
}
