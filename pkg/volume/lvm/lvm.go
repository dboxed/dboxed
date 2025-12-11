package lvm

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"

	"github.com/dboxed/dboxed/pkg/util/command_helper"
)

type PVEntry struct {
	PvName string     `json:"pv_name"`
	VgName string     `json:"vg_name"`
	PvFmt  string     `json:"pv_fmt"`
	PvAttr string     `json:"pv_attr"`
	PvSize string     `json:"pv_size"`
	PvFree string     `json:"pv_free"`
	PvTags stringList `json:"pv_tags"`
}

type pvsReport struct {
	Report []struct {
		Pv []PVEntry `json:"pv"`
	} `json:"report"`
}

type VGEntry struct {
	VgName    string     `json:"vg_name"`
	PvCount   string     `json:"pv_count"`
	LvCount   string     `json:"lv_count"`
	SnapCount string     `json:"snap_count"`
	VgAttr    string     `json:"vg_attr"`
	VgSize    string     `json:"vg_size"`
	VgFree    string     `json:"vg_free"`
	VgTags    stringList `json:"pv_tags"`
}
type vgsReport struct {
	Report []struct {
		Vg []VGEntry `json:"vg"`
	} `json:"report"`
}

type LVEntry struct {
	LvName          string `json:"lv_name"`
	VgName          string `json:"vg_name"`
	LvFullName      string `json:"lv_full_name"`
	LvAttr          string `json:"lv_attr"`
	LvSize          string `json:"lv_size"`
	PoolLv          string `json:"pool_lv"`
	Origin          string `json:"origin"`
	DataPercent     string `json:"data_percent"`
	MetadataPercent string `json:"metadata_percent"`
	MovePv          string `json:"move_pv"`
	MirrorLog       string `json:"mirror_log"`
	CopyPercent     string `json:"copy_percent"`
	ConvertLv       string `json:"convert_lv"`

	LvTags   stringList `json:"lv_tags"`
	LvActive string     `json:"lv_active"`
}

type stringList struct {
	L []string
}

func (sl *stringList) UnmarshalJSON(b []byte) error {
	var s string
	err := json.Unmarshal(b, &s)
	if err != nil {
		return err
	}
	sl.L = strings.Split(s, ",")
	return nil
}

type lvsReport struct {
	Report []struct {
		Lv []LVEntry `json:"lv"`
	} `json:"report"`
}

func buildColNames[T any]() []string {
	var ret []string
	t := reflect.TypeFor[T]()
	for i := range t.NumField() {
		f := t.Field(i)
		jt := f.Tag.Get("json")
		if jt != "" {
			ret = append(ret, jt)
		}
	}
	return ret
}

func ListPVs(ctx context.Context) ([]PVEntry, error) {
	h, err := command_helper.RunCommandJson[pvsReport](ctx, "pvs", "--reportformat=json", "-o", strings.Join(buildColNames[PVEntry](), ","))
	if err != nil {
		return nil, err
	}
	return h.Report[0].Pv, nil
}

func ListVGs(ctx context.Context) ([]VGEntry, error) {
	h, err := command_helper.RunCommandJson[vgsReport](ctx, "vgs", "--reportformat=json", "-o", strings.Join(buildColNames[VGEntry](), ","))
	if err != nil {
		return nil, err
	}
	return h.Report[0].Vg, nil
}

func ListLVs(ctx context.Context) ([]LVEntry, error) {
	h, err := command_helper.RunCommandJson[lvsReport](ctx, "lvs", "--all", "--reportformat=json", "-o", strings.Join(buildColNames[LVEntry](), ","))
	if err != nil {
		return nil, err
	}
	return h.Report[0].Lv, nil
}

func PVCreate(ctx context.Context, dev string) error {
	err := command_helper.RunCommand(ctx, "pvcreate", dev)
	if err != nil {
		return err
	}
	return nil
}

func PVAddTags(ctx context.Context, dev string, tags []string) error {
	var args []string
	for _, t := range tags {
		args = append(args, "--addtag", t)
	}
	args = append(args, dev)
	err := command_helper.RunCommand(ctx, "pvchange", args...)
	if err != nil {
		return err
	}
	return nil
}

func VGCreate(ctx context.Context, vgName string, devs []string, tags []string) error {
	args := []string{
		"--setautoactivation", "n",
		vgName,
	}
	args = append(args, devs...)
	for _, t := range tags {
		args = append(args, "--addtag", t)
	}
	err := command_helper.RunCommand(ctx, "vgcreate", args...)
	if err != nil {
		return err
	}
	return nil
}

func VGGet(ctx context.Context, vgName string) (*VGEntry, error) {
	vgs, err := ListVGs(ctx)
	if err != nil {
		return nil, err
	}
	for _, vg := range vgs {
		if vg.VgName == vgName {
			return &vg, nil
		}
	}
	return nil, os.ErrNotExist
}

func VGDeactivate(ctx context.Context, vgName string) error {
	cmd := command_helper.CommandHelper{
		Command: "vgchange",
		Args: []string{
			"-an",
			"--ignoremonitoring",
			vgName,
		},
		Env: []string{
			fmt.Sprintf("LVM_SUPPRESS_FD_WARNINGS=1"),
		},
	}
	err := cmd.Run(ctx)
	if err != nil {
		return err
	}
	return nil
}

func LVGet(ctx context.Context, vgName string, lvName string) (*LVEntry, error) {
	lvs, err := ListLVs(ctx)
	if err != nil {
		return nil, err
	}
	for _, lv := range lvs {
		if lv.VgName == vgName && lv.LvName == lvName {
			return &lv, nil
		}
	}
	return nil, os.ErrNotExist
}

func LVCreate(ctx context.Context, vgName string, lvName string, size int64, tags []string) error {
	args := []string{
		"--name", lvName,
		"-L", fmt.Sprintf("%dB", size),
		"--setautoactivation", "n",
		vgName,
	}
	for _, t := range tags {
		args = append(args, "--addtag", t)
	}
	err := command_helper.RunCommand(ctx, "lvcreate", args...)
	if err != nil {
		return err
	}
	return nil
}

func LVSnapCreate100Free(ctx context.Context, vgName string, lvName string, snapName string) error {
	args := []string{
		"-l100%FREE",
		"-s", "--name", snapName,
		fmt.Sprintf("%s/%s", vgName, lvName),
	}
	err := command_helper.RunCommand(ctx, "lvcreate", args...)
	if err != nil {
		return err
	}
	return nil
}

func LVRemove(ctx context.Context, vgName string, lvName string) error {
	args := []string{
		"-f",
		fmt.Sprintf("%s/%s", vgName, lvName),
	}
	err := command_helper.RunCommand(ctx, "lvremove", args...)
	if err != nil {
		return err
	}
	return nil
}

func LVActivate(ctx context.Context, vgName string, lvName string, activate bool) error {
	return LVActivateFullName(ctx, fmt.Sprintf("%s/%s", vgName, lvName), activate)
}

func LVActivateFullName(ctx context.Context, fullName string, activate bool) error {
	args := []string{
		"-K",
		"-y",
	}
	if activate {
		args = append(args, "-ay")
	} else {
		args = append(args, "-an")
	}
	args = append(args, "--ignoremonitoring")
	args = append(args, fullName)
	err := command_helper.RunCommand(ctx, "lvchange", args...)
	if err != nil {
		return err
	}
	return nil
}

func FindLVsWithTag(ctx context.Context, tag string) ([]LVEntry, error) {
	lvs, err := ListLVs(ctx)
	if err != nil {
		return nil, err
	}

	var ret []LVEntry
	for _, lv := range lvs {
		if slices.Contains(lv.LvTags.L, tag) {
			ret = append(ret, lv)
		}
	}
	return ret, nil
}

func BuildDevPath(vgName string, lvName string, evalSymlinks bool) (string, error) {
	p := filepath.Join("/dev", vgName, lvName)
	if evalSymlinks {
		var err error
		p, err = filepath.EvalSymlinks(p)
		if err != nil {
			return "", err
		}
	}
	return p, nil
}
