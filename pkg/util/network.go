package util

import (
	"fmt"
	"runtime"

	"github.com/vishvananda/netns"
)

func NewNetNsWithoutEnter(name string) (netns.NsHandle, error) {
	origns, err := netns.Get()
	if err != nil {
		return 0, err
	}

	ns, err := netns.NewNamed(name)
	if err != nil {
		return 0, err
	}
	err = netns.Set(origns)
	if err != nil {
		panic(fmt.Errorf("failed to set back origns, which is not recovarable: %w", err))
	}
	return ns, nil
}

func RunInNetNs(ns netns.NsHandle, fn func() error) error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	o, err := netns.Get()
	if err != nil {
		return err
	}

	if o == ns {
		return fn()
	}

	err = netns.Set(ns)
	if err != nil {
		return err
	}
	defer func() {
		err := netns.Set(o)
		if err != nil {
			panic(err)
		}
	}()

	err = fn()
	if err != nil {
		return err
	}
	return nil
}

func RunInNetNsOptional(ns *netns.NsHandle, fn func() error) error {
	if ns == nil {
		return fn()
	}
	return RunInNetNs(*ns, fn)
}
