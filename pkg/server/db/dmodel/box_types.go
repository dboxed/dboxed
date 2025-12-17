package dmodel

type BoxType string

const (
	BoxTypeNormal       BoxType = "normal"
	BoxTypeGitSpec      BoxType = "git-spec"
	BoxTypeLoadBalancer BoxType = "load-balancer"
)
