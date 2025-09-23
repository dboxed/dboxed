package aws

import (
	"context"
	"encoding/base64"
	"errors"
	"log/slog"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/smithy-go"
	"github.com/dboxed/dboxed/pkg/server/cloud_utils"
	"github.com/dboxed/dboxed/pkg/server/db/dmodel"
	"github.com/dboxed/dboxed/pkg/server/db/querier"
	"golang.org/x/crypto/ssh"
)

func (r *Reconciler) reconcileSshKey(ctx context.Context) error {
	q := querier.GetQuerier(ctx)

	if r.mp.SshKeyPublic == nil {
		if r.mp.HasFinalizer("aws-ssh-key") {
			return r.deleteSshKeyPair(ctx)
		}
		return nil
	}

	keyName := cloud_utils.BuildAwsSshKeyName(ctx, r.mp.Name, r.mp.ID)

	pk, _, _, _, err := ssh.ParseAuthorizedKey([]byte(*r.mp.SshKeyPublic))
	if err != nil {
		return err
	}
	fingerprint := ssh.FingerprintSHA256(pk)

	resp, err := r.ec2Client.DescribeKeyPairs(ctx, &ec2.DescribeKeyPairsInput{
		KeyNames: []string{keyName},
	})
	if err != nil {
		var err2 *smithy.GenericAPIError
		if errors.As(err, &err2) && err2.Code == "InvalidKeyPair.NotFound" {
			return r.createSshKeyPair(ctx)
		} else {
			return err
		}
	}

	// AWS is using standard encoding while ssh.FingerprintSHA256 uses raw encoding
	fingerprint2, err := base64.StdEncoding.DecodeString(*resp.KeyPairs[0].KeyFingerprint)
	if err != nil {
		return err
	}
	fingerprint3 := "SHA256:" + base64.RawStdEncoding.EncodeToString(fingerprint2)
	if fingerprint == fingerprint3 {
		return dmodel.AddFinalizer(q, r.mp, "aws-ssh-key")
	}

	err = r.deleteSshKeyPair(ctx)
	if err != nil {
		return err
	}
	return r.createSshKeyPair(ctx)
}

func (r *Reconciler) createSshKeyPair(ctx context.Context) error {
	q := querier.GetQuerier(ctx)
	keyName := cloud_utils.BuildAwsSshKeyName(ctx, r.mp.Name, r.mp.ID)

	var tags []types.Tag
	for k, v := range cloud_utils.BuildCloudBaseTags(r.mp.ID, r.mp.WorkspaceID) {
		tags = append(tags, types.Tag{Key: &k, Value: &v})
	}

	r.log.InfoContext(ctx, "adding ssh key", slog.Any("sshKeyName", keyName))
	_, err := r.ec2Client.ImportKeyPair(ctx, &ec2.ImportKeyPairInput{
		KeyName:           &keyName,
		PublicKeyMaterial: []byte(*r.mp.SshKeyPublic),
		TagSpecifications: []types.TagSpecification{
			{ResourceType: types.ResourceTypeKeyPair, Tags: tags},
		},
	})
	if err != nil {
		return err
	}

	return dmodel.AddFinalizer(q, r.mp, "aws-ssh-key")
}

func (r *Reconciler) deleteSshKeyPair(ctx context.Context) error {
	q := querier.GetQuerier(ctx)
	keyName := cloud_utils.BuildAwsSshKeyName(ctx, r.mp.Name, r.mp.ID)

	r.log.InfoContext(ctx, "deleting ssh key", slog.Any("sshKeyName", keyName))
	_, err := r.ec2Client.DeleteKeyPair(ctx, &ec2.DeleteKeyPairInput{
		KeyName: &keyName,
	})
	if err != nil {
		var err2 *smithy.GenericAPIError
		if !errors.As(err, &err2) || err2.Code != "InvalidKeyPair.NotFound" {
			return err
		}
	}

	return dmodel.RemoveFinalizer(q, r.mp, "aws-ssh-key")
}
