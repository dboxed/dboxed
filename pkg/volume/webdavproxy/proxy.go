package webdavproxy

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"

	"golang.org/x/net/webdav"
)

type Proxy struct {
	fs         webdav.FileSystem
	listenAddr string

	httpServer *http.Server
}

func NewProxy(fs webdav.FileSystem, listenAddr string) (*Proxy, error) {
	p := &Proxy{
		fs:         fs,
		listenAddr: listenAddr,
	}
	return p, nil
}

func (p *Proxy) Start(ctx context.Context) (net.Addr, error) {
	listenAddr, err := net.ResolveTCPAddr("tcp", p.listenAddr)
	if err != nil {
		return nil, err
	}

	l, err := net.ListenTCP("tcp", listenAddr)
	if err != nil {
		return nil, err
	}

	slog.Info("starting webdav proxy", slog.Any("listenAddr", l.Addr().String()))
	p.httpServer = &http.Server{
		Handler: p.handler(),
		BaseContext: func(listener net.Listener) context.Context {
			return ctx
		},
	}
	go func() {
		err := p.httpServer.Serve(l)
		if err != nil {
			if !errors.Is(err, http.ErrServerClosed) {
				slog.Error("webdav proxy server serve exited with error", slog.Any("error", err))
				return
			}
		}
		slog.Info("webdav proxy stopped")
	}()
	return l.Addr(), nil
}

func (p *Proxy) Stop() error {
	return p.httpServer.Close()
}

func (p *Proxy) handler() http.Handler {
	wh := &webdav.Handler{
		FileSystem: p.fs,
		LockSystem: webdav.NewMemLS(),
	}

	wrapper := http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Method == "HEAD" || request.Method == "GET" {
			// we're not interested in returning the content-type. the peeking that the webdav server does for this
			// slows down the whole process.
			writer.Header().Set("Content-Type", "")
		}
		wh.ServeHTTP(writer, request)
	})

	return wrapper
}
