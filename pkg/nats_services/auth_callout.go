package nats_services

import (
	"context"
	"encoding/base64"
	"fmt"
	"log/slog"

	"github.com/dboxed/dboxed/pkg/server/config"
	"github.com/dboxed/dboxed/pkg/server/db/dmodel"
	querier2 "github.com/dboxed/dboxed/pkg/server/db/querier"
	"github.com/dboxed/dboxed/pkg/server/nats_utils"
	"github.com/nats-io/jwt/v2"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/micro"
	"github.com/nats-io/nkeys"
)

type AuthCalloutService struct {
	config config.Config

	nc            *nats.Conn
	issuerKeyPair nkeys.KeyPair
	xkeyKeyPair   nkeys.KeyPair

	servicesNkey string
}

func NewAuthCalloutService(config config.Config) (*AuthCalloutService, error) {
	authKeyPair, authPub, err := loadSeed(config.Nats.AuthUserSeed)
	if err != nil {
		return nil, err
	}
	issuerKeyPair, _, err := loadSeed(config.Nats.AuthIssuerSeed)
	if err != nil {
		return nil, err
	}
	xkeyKeyPair, _, err := loadSeed(config.Nats.AuthEncryptionSeed)
	if err != nil {
		return nil, err
	}
	_, servicesNkey, err := loadSeed(config.Nats.ServicesSeed)
	if err != nil {
		return nil, err
	}

	nc, err := nats.Connect(config.Nats.Url,
		nats.Nkey(authPub, authKeyPair.Sign),
		nats.ConnectHandler(func(conn *nats.Conn) {
			slog.Debug("nats connected for auth callout service")
		}),
		nats.DisconnectErrHandler(func(conn *nats.Conn, err error) {
			slog.Debug("nats disconnected for auth callout service", slog.Any("error", err))
		}),
		nats.ReconnectHandler(func(conn *nats.Conn) {
			slog.Debug("nats reconnected for auth callout service")
		}),
		nats.ReconnectErrHandler(func(conn *nats.Conn, err error) {
			slog.Debug("nats reconnect failed for auth callout service", slog.Any("error", err))
		}),
	)
	if err != nil {
		return nil, err
	}
	doClose := true
	defer func() {
		if doClose {
			nc.Close()
		}
	}()

	s := &AuthCalloutService{
		config:        config,
		nc:            nc,
		issuerKeyPair: issuerKeyPair,
		xkeyKeyPair:   xkeyKeyPair,
		servicesNkey:  servicesNkey,
	}

	doClose = false
	return s, nil
}

func (s *AuthCalloutService) Run(ctx context.Context) error {
	// Create a service for auth callout with an endpoint binding to
	// the required subject. This allows for running multiple instances
	// to distribute the load, observe stats, and provide high availability.
	srv, err := micro.AddService(s.nc, micro.Config{
		Name:        "auth-callout",
		Version:     "0.0.1",
		Description: "Auth callout service.",
	})
	if err != nil {
		return err
	}

	g := srv.
		AddGroup("$SYS").
		AddGroup("REQ").
		AddGroup("USER")

	err = g.AddEndpoint("AUTH", micro.HandlerFunc(func(req micro.Request) {
		s.msgHandler(ctx, req)
	}))
	if err != nil {
		return err
	}

	<-ctx.Done()

	return nil
}

func (s *AuthCalloutService) msgHandler(ctx context.Context, req micro.Request) {
	rc, err := s.decodeRequest(req)
	if err != nil {
		slog.ErrorContext(ctx, "failed to decode message", slog.Any("error", err))
		_ = req.Respond(nil)
		return
	}
	resc, err := s.handleReq(ctx, rc)
	if err != nil {
		slog.ErrorContext(ctx, "failed to handle request", slog.Any("error", err))
		resc = jwt.NewAuthorizationResponseClaims(rc.UserNkey)
		resc.Audience = rc.Server.ID
		resc.Error = err.Error()
	}
	resb, err := s.encodeResponseMsg(req, resc)
	if err != nil {
		slog.ErrorContext(ctx, "failed encode response", slog.Any("error", err))
		resb = nil
	}
	err = req.Respond(resb)
	if err != nil {
		slog.ErrorContext(ctx, "failed sending response", slog.Any("error", err))
	}
}

func (s *AuthCalloutService) encodeResponseMsg(req micro.Request, rc *jwt.AuthorizationResponseClaims) ([]byte, error) {
	token, err := rc.Encode(s.issuerKeyPair)
	if err != nil {
		return nil, fmt.Errorf("error encoding response JWT: %w", err)
	}

	data := []byte(token)

	serverXkey := req.Headers().Get("Nats-Server-Xkey")
	if len(serverXkey) > 0 {
		data, err = s.xkeyKeyPair.Seal(data, serverXkey)
		if err != nil {
			return nil, fmt.Errorf("error encrypting response JWT: %w", err)
		}
	}

	return data, nil
}

func (s *AuthCalloutService) decodeRequest(req micro.Request) (*jwt.AuthorizationRequestClaims, error) {
	var token []byte

	// Check for Xkey header and decrypt
	serverXkey := req.Headers().Get("Nats-Server-Xkey")
	if serverXkey == "" {
		return nil, fmt.Errorf("message must be encrypted")
	}

	// Decrypt the message.
	var err error
	token, err = s.xkeyKeyPair.Open(req.Data(), serverXkey)
	if err != nil {
		return nil, fmt.Errorf("error decrypting message")
	}

	// Decode the authorization request claims.
	rc, err := jwt.DecodeAuthorizationRequestClaims(string(token))
	if err != nil {
		return nil, fmt.Errorf("error while decoding claims: %w", err)
	}
	return rc, nil
}

func (s *AuthCalloutService) handleReq(ctx context.Context, rc *jwt.AuthorizationRequestClaims) (*jwt.AuthorizationResponseClaims, error) {
	var uc *jwt.UserClaims
	var err error
	if rc.ConnectOptions.Nkey != "" {
		nkeyPub, err := nkeys.FromPublicKey(rc.ConnectOptions.Nkey)
		if err != nil {
			return nil, err
		}

		sig, err := base64.RawURLEncoding.DecodeString(rc.ConnectOptions.SignedNonce)
		if err != nil {
			return nil, err
		}
		err = nkeyPub.Verify([]byte(rc.ClientInformation.Nonce), sig)
		if err != nil {
			return nil, err
		}

		uc, err = s.buildAdminUser(ctx, rc)
		if err != nil {
			return nil, err
		}
		if uc == nil {
			uc, err = s.buildServicesUser(ctx, rc)
			if err != nil {
				return nil, err
			}
		}
		if uc == nil {
			uc, err = s.buildWorkspaceUser(ctx, rc)
			if err != nil {
				return nil, err
			}
		}
	}

	if uc == nil {
		uc, err = s.buildBoxUser(ctx, rc)
		if err != nil {
			return nil, err
		}
	}
	if uc == nil {
		return nil, fmt.Errorf("unknown nkey/token")
	}

	// Validate the claims.
	vr := jwt.CreateValidationResults()
	uc.Validate(vr)
	if len(vr.Errors()) > 0 {
		return nil, fmt.Errorf("error validating claims")
	}

	// Sign it with the issuer key since this is non-operator mode.
	ejwt, err := uc.Encode(s.issuerKeyPair)
	if err != nil {
		return nil, fmt.Errorf("error signing user JWT")
	}

	resc := jwt.NewAuthorizationResponseClaims(rc.UserNkey)
	resc.Audience = rc.Server.ID
	resc.Jwt = ejwt

	return resc, nil
}

func (s *AuthCalloutService) buildAdminUser(ctx context.Context, rc *jwt.AuthorizationRequestClaims) (*jwt.UserClaims, error) {
	if rc.ConnectOptions.Nkey != s.config.Nats.AdminNkey {
		return nil, nil
	}

	uc := jwt.NewUserClaims(rc.UserNkey)
	uc.Name = fmt.Sprintf("admin")
	uc.Audience = "DBOXED"
	uc.Permissions = jwt.Permissions{
		Sub: jwt.Permission{
			Allow: []string{">"},
		},
		Pub: jwt.Permission{
			Allow: []string{">"},
		},
	}
	return uc, nil
}

func (s *AuthCalloutService) buildServicesUser(ctx context.Context, rc *jwt.AuthorizationRequestClaims) (*jwt.UserClaims, error) {
	if rc.ConnectOptions.Nkey != s.servicesNkey {
		return nil, nil
	}

	uc := jwt.NewUserClaims(rc.UserNkey)
	uc.Name = fmt.Sprintf("services")
	uc.Audience = "DBOXED"
	uc.Permissions = jwt.Permissions{
		Sub: jwt.Permission{
			Allow: []string{
				"_INBOX.>",
				"$SRV.PING",
				"$SRV.PING.dboxed",
				"$SRV.PING.dboxed.>",
				"$SRV.STATS",
				"$SRV.STATS.dboxed",
				"$SRV.STATS.dboxed.>",
				"$SRV.INFO",
				"$SRV.INFO.dboxed",
				"$SRV.INFO.dboxed.>",
				"dboxed.get-box-spec.>",
			},
		},
		Pub: jwt.Permission{
			//Allow: []string{""},
		},
	}
	return uc, nil
}

func (s *AuthCalloutService) buildWorkspaceUser(ctx context.Context, rc *jwt.AuthorizationRequestClaims) (*jwt.UserClaims, error) {
	if rc.ConnectOptions.Nkey == "" {
		return nil, nil
	}

	q := querier2.GetQuerier(ctx)
	w, err := dmodel.GetWorkspaceByNkey(q, rc.ConnectOptions.Nkey)
	if err != nil {
		if querier2.IsSqlNotFoundError(err) {
			return nil, nil
		}
		return nil, err
	}

	uc := jwt.NewUserClaims(rc.UserNkey)
	uc.Name = fmt.Sprintf("workspace-%d-%s", w.ID, w.Name)
	uc.Audience = "DBOXED"
	uc.Permissions = s.buildWorkspacePermissions(ctx, *w)

	return uc, nil
}

func (s *AuthCalloutService) buildBoxUser(ctx context.Context, rc *jwt.AuthorizationRequestClaims) (*jwt.UserClaims, error) {
	q := querier2.GetQuerier(ctx)
	var box *dmodel.Box
	var err error
	if rc.ConnectOptions.Nkey != "" {
		box, err = dmodel.GetBoxByNkey(q, rc.ConnectOptions.Nkey)
		if err != nil {
			if querier2.IsSqlNotFoundError(err) {
				return nil, nil
			}
			return nil, err
		}
	} else if rc.ConnectOptions.Token != "" {
		token, err := dmodel.GetTokenByToken(q, rc.ConnectOptions.Token)
		if err != nil {
			return nil, err
		}
		if token.BoxID == nil {
			return nil, nil
		}
		box, err = dmodel.GetBoxById(q, &token.WorkspaceID, *token.BoxID, true)
		if err != nil {
			return nil, err
		}
	}

	if box == nil {
		return nil, nil
	}

	uc := jwt.NewUserClaims(rc.UserNkey)
	uc.Name = fmt.Sprintf("box-%d-%s", box.ID, box.Name)
	uc.Audience = "DBOXED"
	uc.Permissions = s.buildBoxPermissions(ctx, *box)

	return uc, nil
}

func (s *AuthCalloutService) buildWorkspacePermissions(ctx context.Context, w dmodel.Workspace) jwt.Permissions {
	boxSpecsKvName := nats_utils.BuildBoxSpecsKVStoreName(ctx, w.ID)
	metadataKvName := nats_utils.BuildMetadataKVStoreName(ctx, w.ID)
	logsStreamName := nats_utils.BuildLogsStreamName(ctx, w.ID)

	return jwt.Permissions{
		Sub: jwt.Permission{
			Allow: []string{
				"_INBOX.>",
				fmt.Sprintf("%s.>", logsStreamName),
			},
		},
		Pub: jwt.Permission{
			Allow: []string{
				"$JS.API.INFO",
				"$JS.API.STREAM.LIST",

				fmt.Sprintf("$JS.API.STREAM.*.KV_%s", boxSpecsKvName),
				fmt.Sprintf("$JS.API.CONSUMER.*.KV_%s.>", boxSpecsKvName),
				fmt.Sprintf("$JS.API.DIRECT.*.KV_%s.>", boxSpecsKvName),
				fmt.Sprintf("$KV.%s.>", boxSpecsKvName),

				fmt.Sprintf("$JS.API.STREAM.*.KV_%s", metadataKvName),
				fmt.Sprintf("$JS.API.CONSUMER.*.KV_%s.>", metadataKvName),
				fmt.Sprintf("$JS.API.DIRECT.*.KV_%s.>", metadataKvName),
				fmt.Sprintf("$KV.%s.>", metadataKvName),

				fmt.Sprintf("$JS.API.STREAM.*.%s", logsStreamName),
				fmt.Sprintf("$JS.API.CONSUMER.*.%s.*", logsStreamName),
				fmt.Sprintf("$JS.API.CONSUMER.MSG.NEXT.%s.*", logsStreamName),
				fmt.Sprintf("$JS.API.CONSUMER.INFO.%s.*", logsStreamName),
				fmt.Sprintf("$JS.ACK.%s.>", logsStreamName),
			},
		},
	}
}

func (s *AuthCalloutService) buildBoxPermissions(ctx context.Context, box dmodel.Box) jwt.Permissions {
	boxSpecsKvName := nats_utils.BuildBoxSpecsKVStoreName(ctx, box.WorkspaceID)
	metadataKvName := nats_utils.BuildMetadataKVStoreName(ctx, box.WorkspaceID)
	logsStreamName := nats_utils.BuildLogsStreamName(ctx, box.WorkspaceID)
	boxSpecSubject := fmt.Sprintf("$KV.%s.box-spec-%d", boxSpecsKvName, box.ID)
	metadataSubject := fmt.Sprintf("$KV.%s.box-%d", metadataKvName, box.ID)
	boxLogsSubject := nats_utils.BuildLogsSubjectName(ctx, box.WorkspaceID, box.ID, "*")

	return jwt.Permissions{
		Sub: jwt.Permission{
			Allow: []string{
				"_INBOX.>",
			},
		},
		Pub: jwt.Permission{
			Allow: []string{
				"$JS.API.INFO",
				"$JS.API.STREAM.LIST",

				fmt.Sprintf("$JS.API.STREAM.INFO.KV_%s", boxSpecsKvName),
				fmt.Sprintf("$JS.API.DIRECT.GET.KV_%s.%s", boxSpecsKvName, boxSpecSubject),
				fmt.Sprintf("$JS.API.CONSUMER.*.KV_%s.*.%s", boxSpecsKvName, boxSpecSubject),
				fmt.Sprintf("$JS.API.CONSUMER.*.KV_%s.*.%s.*", boxSpecsKvName, boxSpecSubject),

				fmt.Sprintf("$JS.API.STREAM.INFO.KV_%s", metadataKvName),
				fmt.Sprintf("$JS.API.DIRECT.GET.KV_%s.%s", metadataKvName, metadataSubject),
				fmt.Sprintf("$JS.API.CONSUMER.*.KV_%s.*.%s", metadataKvName, metadataSubject),
				fmt.Sprintf("$JS.API.CONSUMER.*.KV_%s.*.%s.*", metadataKvName, metadataSubject),
				fmt.Sprintf(fmt.Sprintf("%s.>", metadataSubject)),

				fmt.Sprintf("$JS.API.STREAM.INFO.%s", logsStreamName),
				fmt.Sprintf(boxLogsSubject),
			},
		},
	}
}
