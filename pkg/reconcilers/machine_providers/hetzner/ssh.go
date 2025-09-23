package hetzner

import (
	"context"
	"fmt"

	"github.com/dboxed/dboxed/pkg/server/cloud_utils"
	"github.com/dboxed/dboxed/pkg/server/config"
	"github.com/hetznercloud/hcloud-go/v2/hcloud"
	"golang.org/x/crypto/ssh"
)

func (r *Reconciler) reconcileSshKey(ctx context.Context) error {
	config := config.GetConfig(ctx)
	if r.mp.SshKeyPublic == nil {
		r.sshKeyId = -1
		return nil
	}
	keyName := fmt.Sprintf("%s--%s-%d", config.InstanceName, r.mp.Name, r.mp.ID)

	pk, _, _, _, err := ssh.ParseAuthorizedKey([]byte(*r.mp.SshKeyPublic))
	if err != nil {
		return err
	}
	fingerprint := ssh.FingerprintLegacyMD5(pk)

	k, _, err := r.hcloudClient.SSHKey.GetByFingerprint(ctx, fingerprint)
	if err != nil {
		return err
	}
	if k != nil {
		r.sshKeyId = k.ID
		return nil
	}
	k, _, err = r.hcloudClient.SSHKey.Create(ctx, hcloud.SSHKeyCreateOpts{
		Name:      keyName,
		PublicKey: *r.mp.SshKeyPublic,
		Labels:    cloud_utils.BuildCloudBaseTags(r.mp.ID, r.mp.WorkspaceID),
	})
	if err != nil {
		return err
	}
	r.sshKeyId = k.ID

	return nil
}
