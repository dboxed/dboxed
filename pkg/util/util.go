package util

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"reflect"
	"time"

	"sigs.k8s.io/yaml"
)

func Ptr[T any](v T) *T {
	return &v
}

func MustJson(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return string(b)
}

func CopyViaJson[T any](v T) (ret T, err error) {
	b, err := json.Marshal(v)
	if err != nil {
		return
	}
	err = json.Unmarshal(b, &ret)
	return
}

func MustCopyViaJson[T any](v T) (ret T) {
	c, err := CopyViaJson(v)
	if err != nil {
		panic(err)
	}
	return c
}

func ConvertViaJson[I any, O any](i I) (O, error) {
	var z O
	b, err := json.Marshal(i)
	if err != nil {
		return z, err
	}
	var o O
	err = json.Unmarshal(b, &o)
	if err != nil {
		return z, err
	}
	return o, nil
}

func EqualsViaJson(a any, b any) bool {
	aj := MustJson(a)
	bj := MustJson(b)
	return aj == bj
}

func SleepWithContext(ctx context.Context, d time.Duration) bool {
	select {
	case <-time.After(d):
		return true
	case <-ctx.Done():
		return false
	}
}

func LoopWithPrintErr(ctx context.Context, name string, interval time.Duration, fn func() error) {
	for {
		err := fn()
		if err != nil {
			slog.ErrorContext(ctx, fmt.Sprintf("error in %s", name), slog.Any("error", err))
		}
		if !SleepWithContext(ctx, interval) {
			return
		}
	}
}

func AtomicWriteFileYaml(path string, v any, perm os.FileMode) error {
	b, err := yaml.Marshal(v)
	if err != nil {
		return err
	}
	return AtomicWriteFile(path, b, perm)
}

func AtomicWriteFile(path string, b []byte, perm os.FileMode) error {
	tmpFile, err := os.CreateTemp(filepath.Dir(path), filepath.Base(path)+".*")
	if err != nil {
		return err
	}
	defer tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.Write(b)
	if err != nil {
		return err
	}
	err = tmpFile.Close()
	if err != nil {
		return err
	}

	err = os.Chmod(tmpFile.Name(), perm)
	if err != nil {
		return err
	}

	err = os.Rename(tmpFile.Name(), path)
	if err != nil {
		return fmt.Errorf("failed to rename file: %w", err)
	}
	return nil
}

func IsAnyNil(v any) bool {
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Chan, reflect.Func, reflect.Map, reflect.Pointer, reflect.UnsafePointer, reflect.Interface, reflect.Slice:
		return rv.IsNil()
	default:
		return false
	}
}
