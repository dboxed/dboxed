package nats_conn_pool

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nkeys"
)

type NatsConnectionPool struct {
	ctx       context.Context
	freeConns []*natsConnEntry
	m         sync.Mutex
}

type natsConnEntry struct {
	url    string
	nkey   string
	conn   *nats.Conn
	used   bool
	expire time.Time
}

func NewNatsConnectionPool(ctx context.Context) *NatsConnectionPool {
	p := &NatsConnectionPool{
		ctx: ctx,
	}
	return p
}

func (p *NatsConnectionPool) parseNkeySeed(nkeySeed string) (nkeys.KeyPair, string, error) {
	authKeyPair, err := nkeys.FromSeed([]byte(nkeySeed))
	if err != nil {
		return nil, "", nil
	}
	authPub, err := authKeyPair.PublicKey()
	if err != nil {
		return nil, "", nil
	}
	return authKeyPair, authPub, nil
}

func (p *NatsConnectionPool) Connect(url string, nkeySeed string) (*nats.Conn, error) {
	kp, nkey, err := p.parseNkeySeed(nkeySeed)
	if err != nil {
		return nil, err
	}

	p.m.Lock()

	now := time.Now()
	for i, e := range p.freeConns {
		if e == nil || e.used {
			continue
		}
		if now.After(e.expire) {
			e.conn.Close()
			p.freeConns[i] = nil
			continue
		}
	}
	for _, e := range p.freeConns {
		if e == nil || e.used {
			continue
		}
		if e.url == url && e.nkey == nkey {
			e.used = true
			p.m.Unlock()
			return e.conn, nil
		}
	}

	c, err := nats.Connect(url,
		nats.Nkey(nkey, kp.Sign),
		nats.ConnectHandler(func(conn *nats.Conn) {
			slog.Debug("nats connected for pool")
		}),
		nats.DisconnectErrHandler(func(conn *nats.Conn, err error) {
			slog.Debug("nats disconnected for pool", slog.Any("error", err))
		}),
		nats.ReconnectHandler(func(conn *nats.Conn) {
			slog.Debug("nats reconnected for pool")
		}),
		nats.ReconnectErrHandler(func(conn *nats.Conn, err error) {
			slog.Debug("nats reconnect failed for pool", slog.Any("error", err))
		}),
		nats.ClosedHandler(func(conn *nats.Conn) {
			slog.Debug("nats connection closed for pool")
		}),
	)
	if err != nil {
		p.m.Unlock()
		return nil, err
	}

	defer p.m.Unlock()

	e := &natsConnEntry{
		url:    url,
		nkey:   nkey,
		conn:   c,
		used:   true,
		expire: time.Now().Add(1 * time.Minute),
	}
	needAppend := true
	for i := range p.freeConns {
		if p.freeConns[i] == nil {
			p.freeConns[i] = e
			needAppend = false
			break
		}
	}
	if needAppend {
		p.freeConns = append(p.freeConns, e)
	}
	return c, nil
}

func (p *NatsConnectionPool) Release(c *nats.Conn) {
	p.m.Lock()
	defer p.m.Unlock()
	for _, e := range p.freeConns {
		if e == nil {
			continue
		}
		if e.conn == c {
			e.used = false
			break
		}
	}
}
