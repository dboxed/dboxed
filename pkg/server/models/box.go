package models

import (
	"context"
	"encoding/json"
	"time"

	"github.com/dboxed/dboxed/pkg/boxspec"
	"github.com/dboxed/dboxed/pkg/server/db/dmodel"
	"github.com/dboxed/dboxed/pkg/server/global"
	"github.com/dboxed/dboxed/pkg/server/nats_utils"
	"github.com/dboxed/dboxed/pkg/util"
)

// todo camelcase

type Box struct {
	ID        int64     `json:"id"`
	Workspace int64     `json:"workspace"`
	CreatedAt time.Time `json:"created_at"`

	Uuid string `json:"uuid"`
	Name string `json:"name"`

	Machine *int64 `json:"machine"`

	Network     *int64              `json:"network"`
	NetworkType *global.NetworkType `json:"network_type"`

	DboxedVersion string `json:"dboxed_version"`

	BoxSpec boxspec.BoxSpec `json:"box_spec"`

	BoxUrl string `json:"boxUrl"`
}

type CreateBox struct {
	Name string `json:"name"`

	Network *int64 `json:"network,omitempty"`
}

type UpdateBox struct {
	BoxSpec *boxspec.BoxSpec `json:"boxSpec"`
}

type BoxToken struct {
	Token string `json:"token"`
}

func BoxFromDB(ctx context.Context, s dmodel.Box) (*Box, error) {
	var networkType *global.NetworkType
	if s.NetworkType != nil {
		networkType = util.Ptr(global.NetworkType(*s.NetworkType))
	}
	ret := &Box{
		ID:        s.ID,
		Workspace: s.WorkspaceID,
		CreatedAt: s.CreatedAt,

		Uuid: s.Uuid,
		Name: s.Name,

		Machine: s.MachineID,

		Network:     s.NetworkID,
		NetworkType: networkType,

		DboxedVersion: s.DboxedVersion,

		BoxUrl: nats_utils.BuildBoxSpecsUrl(ctx, s.WorkspaceID, s.ID),
	}

	if len(s.BoxSpec) != 0 {
		bs, err := UnmarshalBoxSpec(s.BoxSpec)
		if err != nil {
			return nil, err
		}
		ret.BoxSpec = *bs
	}

	return ret, nil
}

func MarshalBoxSpec(boxSpec *boxspec.BoxSpec) ([]byte, error) {
	b, err := json.Marshal(boxSpec)
	if err != nil {
		return nil, err
	}
	return util.CompressGzipString(string(b))
}

func UnmarshalBoxSpec(b []byte) (*boxspec.BoxSpec, error) {
	uncompressed, err := util.DecompressGzipString(b)
	if err != nil {
		return nil, err
	}
	var ret boxspec.BoxSpec
	err = json.Unmarshal([]byte(uncompressed), &ret)
	if err != nil {
		return nil, err
	}
	return &ret, nil
}
