package nats_services

import (
	"context"
	"log/slog"
	"time"

	"github.com/dboxed/dboxed/pkg/server/config"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/micro"
)

type DboxedServices struct {
	nc *nats.Conn
}

func NewDboxedServices(ctx context.Context, config config.Config) (*DboxedServices, error) {
	kp, nkey, err := loadSeed(config.Nats.ServicesSeed)
	if err != nil {
		return nil, err
	}

	var nc *nats.Conn
	for {
		nc, err = nats.Connect(config.Nats.Url, nats.Nkey(nkey, kp.Sign))
		if err != nil {
			slog.ErrorContext(ctx, "error while connecting to nats", slog.Any("error", err))
		} else {
			break
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(time.Second):
		}
	}
	doClose := true
	defer func() {
		if doClose {
			nc.Close()
		}
	}()

	s := &DboxedServices{
		nc: nc,
	}

	doClose = false
	return s, nil
}

func (s *DboxedServices) Run(ctx context.Context) error {
	// Create a service for auth callout with an endpoint binding to
	// the required subject. This allows for running multiple instances
	// to distribute the load, observe stats, and provide high availability.
	srv, err := micro.AddService(s.nc, micro.Config{
		Name:    "dboxed",
		Version: "0.0.1",
	})
	if err != nil {
		return err
	}

	g := srv.AddGroup("dboxed")
	_ = g

	// TODO do we actually need services here?
	// if not, we should remove all this

	<-ctx.Done()

	return nil
}
