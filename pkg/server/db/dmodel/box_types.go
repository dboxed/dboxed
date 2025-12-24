package dmodel

type BoxType string

const (
	BoxTypeNormal       BoxType = "normal"
	BoxTypeDboxedSpec   BoxType = "dboxed-spec"
	BoxTypeLoadBalancer BoxType = "load-balancer"
)
