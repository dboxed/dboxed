package dbutils

import (
	"context"

	"github.com/dboxed/dboxed/pkg/server/db/dmodel"
	querier2 "github.com/dboxed/dboxed/pkg/server/db/querier"
	"github.com/dboxed/dboxed/pkg/util"
	"github.com/hashicorp/go-multierror"
)

func DoAndFindChanged[T querier2.HasId](ctx context.Context, list func() ([]T, error), do func(v T) error) error {
	q := querier2.GetQuerier(ctx)

	l, err := list()
	if err != nil {
		return err
	}
	hashes := map[int64]string{}
	for _, v := range l {
		hash, err := util.Sha256SumJson(v)
		if err != nil {
			return err
		}
		hashes[v.GetId()] = hash
	}

	var retErr error
	for _, v := range l {
		err := do(v)
		if err != nil {
			retErr = multierror.Append(err)
		}
	}

	l, err = list()
	if err != nil {
		retErr = multierror.Append(retErr, err)
		return retErr
	}

	for _, v := range l {
		oldHash, ok := hashes[v.GetId()]
		if !ok {
			continue
		}
		hash, err := util.Sha256SumJson(v)
		if err != nil {
			retErr = multierror.Append(retErr, err)
			continue
		}
		if oldHash != hash {
			err = dmodel.AddChangeTracking[T](q, v)
			if err != nil {
				retErr = multierror.Append(retErr, err)
			}
		}
	}

	return retErr
}
