package network

import (
	"context"
	"errors"
	"fmt"
	"github.com/koobox/unboxed/pkg/util"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
	"log/slog"
	"syscall"
	"time"
)

// some code here is copied from netbird (client/internal/routemanager/systemops/systemops_linux.go)

type ruleParams struct {
	oldPriority    int
	priority       int
	fwmark         uint32
	tableID        int
	family         int
	invert         bool
	suppressPrefix int
	description    string
}

func getSetupRules() []ruleParams {
	return []ruleParams{
		// netbird orginally used priority 100, but this clashes with cilium
		{100, 105, 0, syscall.RT_TABLE_MAIN, netlink.FAMILY_V4, false, 0, "rule with suppress prefixlen v4"},
		{100, 105, 0, syscall.RT_TABLE_MAIN, netlink.FAMILY_V6, false, 0, "rule with suppress prefixlen v6"},
	}
}

type NetbirdRulesFix struct {
	SandboxNetworkNamespace netns.NsHandle
}

func (n *NetbirdRulesFix) Start(ctx context.Context) error {
	go func() {
		err := util.RunInNetNs(n.SandboxNetworkNamespace, func() error {
			for {
				err := n.fixNetbirdRulesOnce(ctx)
				if err != nil {
					slog.WarnContext(ctx, "error in fixNetbirdRules", slog.Any("error", err))
				}

				if !util.SleepWithContext(ctx, 5*time.Second) {
					return ctx.Err()
				}
			}
		})
		if err != nil {
			slog.ErrorContext(ctx, "error in fixNetbirdRules", slog.Any("error", err))
		}
	}()
	return nil
}

func (n *NetbirdRulesFix) fixNetbirdRulesOnce(ctx context.Context) error {
	for _, r := range getSetupRules() {
		err := removeRule(r)
		if err != nil {
			return err
		}
		err = addRule(r)
		if err != nil {
			return err
		}
	}
	return nil
}

// addRule adds a routing rule to a specific routing table identified by tableID.
func addRule(params ruleParams) error {
	rule := netlink.NewRule()
	rule.Table = params.tableID
	rule.Mark = params.fwmark
	rule.Family = params.family
	rule.Priority = params.priority
	rule.Invert = params.invert
	rule.SuppressPrefixlen = params.suppressPrefix

	if err := netlink.RuleAdd(rule); err != nil && !errors.Is(err, syscall.EEXIST) && !isOpErr(err) {
		return fmt.Errorf("add routing rule: %w", err)
	}

	return nil
}

// removeRule removes a routing rule from a specific routing table identified by tableID.
func removeRule(params ruleParams) error {
	rule := netlink.NewRule()
	rule.Table = params.tableID
	rule.Mark = params.fwmark
	rule.Family = params.family
	rule.Invert = params.invert
	rule.Priority = params.oldPriority
	rule.SuppressPrefixlen = params.suppressPrefix

	if err := netlink.RuleDel(rule); err != nil && !errors.Is(err, syscall.ENOENT) && !isOpErr(err) {
		return fmt.Errorf("remove routing rule: %w", err)
	}

	return nil
}

func isOpErr(err error) bool {
	// EAFTNOSUPPORT when ipv6 is disabled via sysctl, EOPNOTSUPP when disabled in boot options or otherwise not supported
	if errors.Is(err, syscall.EAFNOSUPPORT) || errors.Is(err, syscall.EOPNOTSUPP) {
		//log.Debugf("route operation not supported: %v", err)
		return true
	}

	return false
}
