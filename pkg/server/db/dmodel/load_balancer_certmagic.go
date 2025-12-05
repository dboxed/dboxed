package dmodel

import (
	"time"

	querier2 "github.com/dboxed/dboxed/pkg/server/db/querier"
)

type LoadBalancerCertmagic struct {
	LoadBalancerId string    `db:"load_balancer_id"`
	Key            string    `db:"key"`
	Value          []byte    `db:"value"`
	LastModified   time.Time `db:"last_modified"`
}

func (v *LoadBalancerCertmagic) Create(q *querier2.Querier) error {
	return querier2.Create(q, v)
}

func (v *LoadBalancerCertmagic) CreateOrUpdate(q *querier2.Querier) error {
	return querier2.CreateOrUpdate(q, v, "(load_balancer_id, key)")
}

func DeleteLoadBalancerCertMagic(q *querier2.Querier, lbId string, key string) error {
	return querier2.DeleteOneByFields[LoadBalancerCertmagic](q, map[string]any{
		"load_balancer_id": lbId,
		"key":              key,
	})
}

func ListLoadBalancerCertmagicKeys(q *querier2.Querier, lbId string, prefix string) ([]string, error) {
	query := "select key from load_balancer_certmagic where load_balancer_id = :load_balancer_id and key like :prefix || '%'"

	var ret []string
	err := q.SelectNamed(&ret, query, map[string]any{
		"load_balancer_id": lbId,
		"prefix":           prefix,
	})
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func GetLoadBalancerCertmagicByKey(q *querier2.Querier, lbId string, key string) (*LoadBalancerCertmagic, error) {
	return querier2.GetOne[LoadBalancerCertmagic](q, map[string]any{
		"load_balancer_id": lbId,
		"key":              key,
	})
}

func (v *LoadBalancerCertmagic) UpdateValue(q *querier2.Querier, value []byte) error {
	v.Value = value
	v.LastModified = time.Now()
	return querier2.UpdateOneFromStruct(q, v, "value")
}
