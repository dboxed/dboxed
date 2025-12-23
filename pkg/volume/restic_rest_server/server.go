package restic_rest_server

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"path"
	"slices"
	"strconv"
	"strings"
	"sync"

	"github.com/dboxed/dboxed/pkg/baseclient"
	"github.com/dboxed/dboxed/pkg/clients"
	"github.com/dboxed/dboxed/pkg/server/models"
	"github.com/gin-gonic/gin"
)

type Server struct {
	client  *baseclient.Client
	client2 *clients.S3ProxyClient

	s3BucketId string
	prefix     string
	listenAddr string

	httpServer *http.Server

	listMutex sync.Mutex
	listError error
	objects   map[string]models.S3ObjectInfo
}

func NewServer(client *baseclient.Client, s3BucketId string, prefix string, listenAddr string) (*Server, error) {
	p := &Server{
		client:     client,
		client2:    &clients.S3ProxyClient{Client: client},
		s3BucketId: s3BucketId,
		prefix:     prefix,
		listenAddr: listenAddr,
	}
	return p, nil
}

func (s *Server) Start(ctx context.Context) (net.Addr, error) {
	listenAddr, err := net.ResolveTCPAddr("tcp", s.listenAddr)
	if err != nil {
		return nil, err
	}

	l, err := net.ListenTCP("tcp", listenAddr)
	if err != nil {
		return nil, err
	}

	slog.Info("starting restic rest server", slog.Any("listenAddr", l.Addr().String()))
	s.httpServer = &http.Server{
		Handler: s.handler(),
		BaseContext: func(listener net.Listener) context.Context {
			return ctx
		},
	}
	go func() {
		err := s.httpServer.Serve(l)
		if err != nil {
			if !errors.Is(err, http.ErrServerClosed) {
				slog.Error("restic rest server serve exited with error", slog.Any("error", err))
				return
			}
		}
		slog.Info("restic rest server stopped")
	}()
	return l.Addr(), nil
}

func (s *Server) Stop() error {
	return s.httpServer.Close()
}

func (s *Server) handler() http.Handler {
	g := gin.New()

	g.HEAD("/config", s.restHeadConfig)
	g.GET("/config", s.restGetConfig)
	g.GET("/:type/", s.restListObjects)
	g.HEAD("/:type/:name", s.restHeadObject)
	g.GET("/:type/:name", s.restGetObject)
	g.POST("/:type/:name", s.restPostObject)
	g.DELETE("/:type/:name", s.restDeleteObject)

	return g
}

func (s *Server) buildPath(objectType string, objectName string) string {
	p := path.Join(s.prefix, objectType)
	if objectName != "" {
		if objectType == "data" && len(objectName) > 2 {
			p = path.Join(p, objectName[0:2], objectName)
		} else {
			p = path.Join(p, objectName)
		}
	}
	return p
}

func (s *Server) listObjects(c *gin.Context) map[string]models.S3ObjectInfo {
	s.listMutex.Lock()
	defer s.listMutex.Unlock()

	setError := func(err error) {
		_ = c.Error(err)
		if baseclient.IsNotFound(err) {
			c.Status(http.StatusNotFound)
		} else {
			c.Status(http.StatusInternalServerError)
		}
	}

	if s.listError != nil {
		setError(s.listError)
		return nil
	}
	if s.objects != nil {
		return s.objects
	}

	slog.DebugContext(c, "listObjects - begin")

	res, err := s.client2.S3ProxyListObjects(c, s.s3BucketId, models.S3ProxyListObjectsRequest{
		Prefix:    s.prefix,
		Recursive: true,
	})
	if err != nil {
		s.listError = err
		setError(err)
		return nil
	}

	slog.DebugContext(c, "listObjects - done", "cnt", len(res.Objects))

	s.objects = map[string]models.S3ObjectInfo{}
	for _, o := range res.Objects {
		s.objects[o.Key] = o
	}

	return s.objects
}

var allowedTypes = []string{
	"data",
	"keys",
	"locks",
	"snapshots",
	"index",
	"config",
}

func (s *Server) headObject(c *gin.Context, objectType string, objectName string) {
	if !slices.Contains(allowedTypes, objectType) {
		c.Status(http.StatusNotFound)
		return
	}

	objects := s.listObjects(c)
	if objects == nil {
		return
	}

	p := s.buildPath(objectType, objectName)

	if o, ok := objects[p]; ok {
		slog.DebugContext(c, "headObject", "ok", true, "p", p, "l", o.Size)
		c.Header("Content-Length", fmt.Sprintf("%d", o.Size))
		c.Status(http.StatusOK)
	} else {
		slog.DebugContext(c, "headObject", "ok", false, "p")
		c.Status(http.StatusNotFound)
	}
}

func (s *Server) loadObject(c *gin.Context, objectType string, objectName string) {
	if !slices.Contains(allowedTypes, objectType) {
		c.Status(http.StatusNotFound)
		return
	}

	objects := s.listObjects(c)
	if objects == nil {
		return
	}

	p := s.buildPath(objectType, objectName)
	o, ok := objects[p]
	if !ok {
		slog.DebugContext(c, "loadObject", "ok", false, "p", p, "l", o.Size)

		c.Status(http.StatusNotFound)
		return
	}

	req, err := http.NewRequest("GET", o.PresignedGetUrl, nil)
	if err != nil {
		_ = c.Error(err)
		c.Status(http.StatusInternalServerError)
		return
	}

	rangeHeader := c.GetHeader("Range")
	if rangeHeader != "" {
		req.Header.Set("Range", rangeHeader)
	}

	slog.DebugContext(c, "loadObject - begin", "p", p, "l", o.Size)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		_ = c.Error(err)
		c.Status(http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	c.Header("Content-Type", "binary/octet-stream")
	c.Header("Content-Length", fmt.Sprintf("%d", resp.ContentLength))
	c.Status(resp.StatusCode)

	_, _ = io.Copy(c.Writer, resp.Body)

	slog.DebugContext(c, "loadObject - done", "p", p, "l", o.Size, "status", resp.StatusCode, "cl", resp.ContentLength)
}

func (s *Server) restHeadConfig(c *gin.Context) {
	s.headObject(c, "config", "")
}

func (s *Server) restGetConfig(c *gin.Context) {
	s.loadObject(c, "config", "")
}

type NameAndSize struct {
	Name string `json:"name"`
	Size int64  `json:"size"`
}

func (s *Server) restListObjects(c *gin.Context) {
	apiVersion := c.GetHeader("Accept")
	isV2 := apiVersion == "application/vnd.x.restic.rest.v2"

	objectType := c.Param("type")
	if !slices.Contains(allowedTypes, objectType) {
		c.Status(http.StatusNotFound)
		return
	}

	objects := s.listObjects(c)
	if objects == nil {
		return
	}

	prefix := s.buildPath(objectType, "")

	var ret []any
	for _, o := range objects {
		if !strings.HasPrefix(o.Key, prefix) {
			continue
		}
		name := path.Base(o.Key)
		if isV2 {
			ret = append(ret, NameAndSize{
				Name: name,
				Size: o.Size,
			})
		} else {
			ret = append(ret, name)
		}
	}

	if isV2 {
		c.Header("Content-Type", "application/vnd.x.restic.rest.v2")
	} else {
		c.Header("Content-Type", "application/vnd.x.restic.rest.v1")
	}

	c.JSON(http.StatusOK, ret)
}

func (s *Server) restHeadObject(c *gin.Context) {
	objectType := c.Param("type")
	objectName := c.Param("name")
	s.headObject(c, objectType, objectName)
}

func (s *Server) restGetObject(c *gin.Context) {
	objectType := c.Param("type")
	objectName := c.Param("name")
	s.loadObject(c, objectType, objectName)
}

func (s *Server) restPostObject(c *gin.Context) {
	objectType := c.Param("type")
	objectName := c.Param("name")

	if !slices.Contains(allowedTypes, objectType) {
		c.Status(http.StatusNotFound)
		return
	}

	contentLengthStr := c.GetHeader("Content-Length")
	contentLength, err := strconv.ParseInt(contentLengthStr, 10, 32)
	if err != nil {
		_ = c.Error(err)
		c.Status(http.StatusInternalServerError)
		return
	}

	p := s.buildPath(objectType, objectName)
	slog.DebugContext(c, "postObject - presign", "p", p, "cl", contentLength)

	res, err := s.client2.S3ProxyPresignPut(c, s.s3BucketId, models.S3ProxyPresignPutRequest{
		Key: p,
	})
	if err != nil {
		if baseclient.IsNotFound(err) {
			c.Status(http.StatusNotFound)
		} else {
			_ = c.Error(err)
			c.Status(http.StatusInternalServerError)
		}
		return
	}

	slog.DebugContext(c, "postObject - begin put", "p", p, "cl", c.GetHeader("Content-Length"))
	req, err := http.NewRequest("PUT", res.PresignedUrl, c.Request.Body)
	if err != nil {
		_ = c.Error(err)
		c.Status(http.StatusInternalServerError)
		return
	}
	req.ContentLength = contentLength

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		_ = c.Error(err)
		c.Status(http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	slog.DebugContext(c, "postObject - done put", "p", p, "cl", c.GetHeader("Content-Length"), "status", resp.StatusCode)

	s.listMutex.Lock()
	defer s.listMutex.Unlock()
	s.objects = nil
	s.listError = nil

	c.Status(resp.StatusCode)
}

func (s *Server) restDeleteObject(c *gin.Context) {
	objectType := c.Param("type")
	objectName := c.Param("name")

	if !slices.Contains(allowedTypes, objectType) {
		c.Status(http.StatusNotFound)
		return
	}

	p := s.buildPath(objectType, objectName)
	_, err := s.client2.S3ProxyDeleteObject(c, s.s3BucketId, models.S3ProxyDeleteObjectRequest{
		Key: p,
	})
	if err != nil {
		if baseclient.IsNotFound(err) {
			c.Status(http.StatusNotFound)
		} else {
			_ = c.Error(err)
			c.Status(http.StatusInternalServerError)
		}
		return
	}

	s.listMutex.Lock()
	defer s.listMutex.Unlock()
	delete(s.objects, p)

	c.Status(http.StatusOK)
}
