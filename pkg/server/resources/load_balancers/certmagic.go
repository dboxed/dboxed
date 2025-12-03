package load_balancers

import (
	"context"
	"encoding/json"
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/dboxed/dboxed/pkg/server/auth_middleware"
	"github.com/dboxed/dboxed/pkg/server/db/dmodel"
	"github.com/dboxed/dboxed/pkg/server/db/querier"
	"github.com/dboxed/dboxed/pkg/server/huma_utils"
)

type lock struct {
	LockTime time.Time `json:"lockTime"`
}

type restPutCertmagicLockInput struct {
	huma_utils.IdByPath
}

var (
	LockExpiration = 2 * time.Minute
)

func (s *LoadBalancerServer) restPutCertmagicLock(c context.Context, i *restPutCertmagicLockInput) (*huma_utils.Empty, error) {
	q := querier.GetQuerier(c)

	err := s.checkLoadBalancerToken(c, i.Id)
	if err != nil {
		return nil, err
	}

	key, err := s.getCertmagicKey(c)
	if err != nil {
		return nil, err
	}
	key = strings.TrimPrefix(key, "/")
	key = path.Join("_locks", key)

	lb, v, err := s.getCertmagicValue(c, i.Id, key)
	if err != nil {
		if !querier.IsSqlNotFoundError(err) {
			return nil, err
		}
	}

	if v != nil {
		var l lock
		err = json.Unmarshal(v.Value, &l)
		if err == nil {
			if !l.LockTime.Add(LockExpiration).Before(time.Now()) {
				return nil, huma.Error409Conflict(fmt.Sprintf("%s is still locked", key), nil)
			}
		}
		err = dmodel.DeleteLoadBalancerCertMagic(q, lb.ID, key)
		if err != nil {
			return nil, err
		}
	}

	l := lock{
		LockTime: time.Now(),
	}
	lBytes, err := json.Marshal(l)
	if err != nil {
		return nil, err
	}
	v = &dmodel.LoadBalancerCertmagic{
		LoadBalancerId: lb.ID,
		Key:            key,
		Value:          lBytes,
		LastModified:   time.Now(),
	}
	err = v.Create(q)
	if err != nil {
		return nil, err
	}

	return &huma_utils.Empty{}, nil
}

func (s *LoadBalancerServer) restDeleteCertmagicLock(c context.Context, i *huma_utils.IdByPath) (*huma_utils.Empty, error) {
	q := querier.GetQuerier(c)
	w := auth_middleware.GetWorkspace(c)

	err := s.checkLoadBalancerToken(c, i.Id)
	if err != nil {
		return nil, err
	}

	key, err := s.getCertmagicKey(c)
	if err != nil {
		return nil, err
	}
	key = strings.TrimPrefix(key, "/")
	key = path.Join("_locks", key)

	lb, err := dmodel.GetLoadBalancerById(q, &w.ID, i.Id, true)
	if err != nil {
		return nil, err
	}

	err = dmodel.DeleteLoadBalancerCertMagic(q, lb.ID, key)
	if err != nil {
		return nil, err
	}

	return &huma_utils.Empty{}, nil
}

type restHeadCertmagicObjectOutput struct {
	Key          string `header:"X-Key"`
	Size         string `header:"X-Size"`
	LastModified string `header:"X-Last-Modified"`
}

func (s *LoadBalancerServer) restHeadCertmagicObject(c context.Context, i *huma_utils.IdByPath) (*restHeadCertmagicObjectOutput, error) {
	err := s.checkLoadBalancerToken(c, i.Id)
	if err != nil {
		return nil, err
	}

	key, err := s.getCertmagicKey(c)
	if err != nil {
		return nil, err
	}
	key = strings.TrimPrefix(key, "/")

	_, v, err := s.getCertmagicValue(c, i.Id, key)
	if err != nil {
		return nil, err
	}

	ret := &restHeadCertmagicObjectOutput{
		Key:          v.Key,
		Size:         fmt.Sprintf("%d", len(v.Value)),
		LastModified: v.LastModified.String(),
	}

	return ret, nil
}

type restGetCertmagicInput struct {
	huma_utils.IdByPath
	Recursive bool `query:"recursive"`
	List      bool `query:"list"`
}

type restGetCertmagicOutput struct {
	Body any
}

func (s *LoadBalancerServer) restGetCertmagicObject(c context.Context, i *restGetCertmagicInput) (*restGetCertmagicOutput, error) {
	err := s.checkLoadBalancerToken(c, i.Id)
	if err != nil {
		return nil, err
	}

	key, err := s.getCertmagicKey(c)
	if err != nil {
		return nil, err
	}
	key = strings.TrimPrefix(key, "/")

	if i.List {
		if !strings.HasSuffix(key, "/") {
			key += "/"
		}

		l, err := s.listCertmagicKeys(c, i.Id, key)
		if err != nil {
			return nil, err
		}
		if i.Recursive {
			return &restGetCertmagicOutput{
				Body: l,
			}, nil
		}
		var lf []string
		for _, k := range lf {
			t := strings.TrimPrefix(k, key)
			if !strings.Contains(t, "/") {
				lf = append(lf, k)
			}
		}
		return &restGetCertmagicOutput{
			Body: lf,
		}, nil
	} else {
		_, v, err := s.getCertmagicValue(c, i.Id, key)
		if err != nil {
			return nil, err
		}
		ret := &restGetCertmagicOutput{
			Body: v.Value,
		}

		return ret, nil
	}
}

type restPutCertmagicInput struct {
	huma_utils.IdByPath
	RawBody []byte `contentType:"application/octet-stream"`
}

func (s *LoadBalancerServer) restPutCertmagicObject(c context.Context, i *restPutCertmagicInput) (*huma_utils.Empty, error) {
	q := querier.GetQuerier(c)
	w := auth_middleware.GetWorkspace(c)

	err := s.checkLoadBalancerToken(c, i.Id)
	if err != nil {
		return nil, err
	}

	key, err := s.getCertmagicKey(c)
	if err != nil {
		return nil, err
	}
	key = strings.TrimPrefix(key, "/")

	lb, err := dmodel.GetLoadBalancerById(q, &w.ID, i.Id, true)
	if err != nil {
		return nil, err
	}

	v := &dmodel.LoadBalancerCertmagic{
		LoadBalancerId: lb.ID,
		Key:            key,
		Value:          i.RawBody,
		LastModified:   time.Now(),
	}
	err = v.CreateOrUpdate(q)
	if err != nil {
		return nil, err
	}

	return &huma_utils.Empty{}, nil
}

func (s *LoadBalancerServer) restDeleteCertmagicObject(c context.Context, i *huma_utils.IdByPath) (*huma_utils.Empty, error) {
	q := querier.GetQuerier(c)
	w := auth_middleware.GetWorkspace(c)

	err := s.checkLoadBalancerToken(c, i.Id)
	if err != nil {
		return nil, err
	}

	key, err := s.getCertmagicKey(c)
	if err != nil {
		return nil, err
	}
	key = strings.TrimPrefix(key, "/")

	lb, err := dmodel.GetLoadBalancerById(q, &w.ID, i.Id, true)
	if err != nil {
		return nil, err
	}

	err = querier.DeleteOneByFields[dmodel.LoadBalancerCertmagic](q, map[string]any{
		"load_balancer_id": lb.ID,
		"key":              key,
	})
	if err != nil {
		return nil, err
	}

	return &huma_utils.Empty{}, nil
}

func (s *LoadBalancerServer) listCertmagicKeys(c context.Context, lbId string, prefix string) ([]string, error) {
	q := querier.GetQuerier(c)
	w := auth_middleware.GetWorkspace(c)

	lb, err := dmodel.GetLoadBalancerById(q, &w.ID, lbId, true)
	if err != nil {
		return nil, err
	}

	l, err := dmodel.ListLoadBalancerCertmagicKeys(q, lb.ID, prefix)
	if err != nil {
		return nil, err
	}

	return l, nil
}

func (s *LoadBalancerServer) getCertmagicKey(c context.Context) (string, error) {
	ginCtx := huma_utils.GetGinContext(c)
	key := ginCtx.Param("key")
	if key == "" {
		return "", huma.Error400BadRequest("missing key")
	}
	return key, nil
}

func (s *LoadBalancerServer) getCertmagicValue(c context.Context, lbId string, key string) (*dmodel.LoadBalancer, *dmodel.LoadBalancerCertmagic, error) {
	q := querier.GetQuerier(c)
	w := auth_middleware.GetWorkspace(c)

	lb, err := dmodel.GetLoadBalancerById(q, &w.ID, lbId, true)
	if err != nil {
		return nil, nil, err
	}

	v, err := dmodel.GetLoadBalancerCertmagicByKey(q, lb.ID, key)
	if err != nil {
		return lb, nil, err
	}

	return lb, v, nil
}

func (s *LoadBalancerServer) checkLoadBalancerToken(c context.Context, lbId string) error {
	token := auth_middleware.GetToken(c)

	if token != nil {
		if token.ForWorkspace {
			return nil
		}
		if token.LoadBalancerId == nil || *token.LoadBalancerId != lbId {
			return huma.Error403Forbidden("no access to load balancer")
		}
	}

	return nil
}
