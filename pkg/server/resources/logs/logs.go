package logs

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/sse"
	"github.com/dboxed/dboxed/pkg/boxspec"
	"github.com/dboxed/dboxed/pkg/server/auth_middleware"
	"github.com/dboxed/dboxed/pkg/server/db/dmodel"
	"github.com/dboxed/dboxed/pkg/server/db/querier"
	"github.com/dboxed/dboxed/pkg/server/huma_utils"
	"github.com/dboxed/dboxed/pkg/server/models"
	"github.com/dboxed/dboxed/pkg/server/resources/huma_metadata"
	"github.com/dboxed/dboxed/pkg/util"
)

type LogsServer struct {
}

func New() *LogsServer {
	return &LogsServer{}
}

func (s *LogsServer) Init(rootGroup huma.API, workspacesGroup huma.API) error {
	allowBoxTokenModifier := huma_utils.MetadataModifier(huma_metadata.AllowBoxToken, true)

	huma.Post(workspacesGroup, "/logs", s.restPostLogs, allowBoxTokenModifier)
	huma.Get(workspacesGroup, "/logs", s.restListLogs)
	sse.Register(workspacesGroup, huma.Operation{
		OperationID: "logs-stream",
		Method:      http.MethodGet,
		Path:        "/logs/{logId}/stream",
		Metadata: map[string]any{
			huma_utils.NoTx: true,
		},
	}, map[string]any{
		"metadata":       models.LogMetadataModel{},
		"logs-batch":     boxspec.LogsBatch{},
		"end-of-history": endOfHistory{},
		"error":          models.LogsError{},
	}, s.sseLogsStream)

	return nil
}

func (s *LogsServer) putLogMetadata(c context.Context, logMetadata boxspec.LogMetadata) (*dmodel.LogMetadata, error) {
	q := querier.GetQuerier(c)
	w := auth_middleware.GetWorkspace(c)

	var metadataBytes []byte
	if logMetadata.Metadata != nil {
		var err error
		metadataBytes, err = json.Marshal(logMetadata.Metadata)
		if err != nil {
			return nil, err
		}
	} else {
		metadataBytes = []byte("{}")
	}

	lm := dmodel.LogMetadata{
		OwnedByWorkspace: dmodel.OwnedByWorkspace{
			WorkspaceID: w.ID,
		},
		FileName: logMetadata.FileName,
		Format:   logMetadata.Format,
		Metadata: string(metadataBytes),
	}
	switch logMetadata.OwnerType {
	case "box":
		lm.BoxID = &logMetadata.OwnerId
	case "machine":
		lm.MachineID = &logMetadata.OwnerId
	default:
		return nil, huma.Error400BadRequest(fmt.Sprintf("unknown owner type %s", logMetadata.OwnerType))
	}
	err := lm.CreateOrUpdate(q)
	if err != nil {
		if querier.IsSqlConstraintViolationError(err) {
			// only viable constraint violation here is when the box id is invalid
			return nil, huma.Error404NotFound("box not found", err)
		}
		return nil, err
	}
	return &lm, nil
}

func (s *LogsServer) restPostLogs(c context.Context, i *huma_utils.JsonBody[models.PostLogs]) (*huma_utils.Empty, error) {
	q := querier.GetQuerier(c)
	w := auth_middleware.GetWorkspace(c)

	switch i.Body.Metadata.OwnerType {
	case "box":
		err := auth_middleware.CheckTokenAccess(c, dmodel.TokenTypeBox, i.Body.Metadata.OwnerId)
		if err != nil {
			return nil, err
		}
	case "machine":
		err := auth_middleware.CheckTokenAccess(c, dmodel.TokenTypeMachine, i.Body.Metadata.OwnerId)
		if err != nil {
			return nil, err
		}
	default:
		return nil, huma.Error400BadRequest(fmt.Sprintf("unknown owner type %s", i.Body.Metadata.OwnerType))
	}

	lm, err := s.putLogMetadata(c, i.Body.Metadata)
	if err != nil {
		return nil, err
	}

	logFileName := i.Body.Metadata.FileName
	if len(logFileName) > 20 {
		logFileName = logFileName[:3] + "..." + logFileName[len(logFileName)-14:]
	}
	huma_utils.ExtraLogValue(c, "lineCount", len(i.Body.Lines))
	huma_utils.ExtraLogValue(c, "logId", lm.ID)
	huma_utils.ExtraLogValue(c, "fileName", logFileName)

	lines := make([]*dmodel.LogLine, 0, len(i.Body.Lines))

	addBytes := int64(0)
	for _, ll := range i.Body.Lines {
		addBytes += int64(len([]byte(ll.Line)))
		dl := &dmodel.LogLine{
			WorkspaceID: w.ID,
			LogID:       lm.ID,
			Line:        ll.Line,
			Time:        ll.Time,
		}
		lines = append(lines, dl)
	}

	huma_utils.ExtraLogValue(c, "addBytes", addBytes)
	if len(i.Body.Lines) == 0 {
		return &huma_utils.Empty{}, nil
	}

	if len(i.Body.Lines) != 0 {
		err = lm.UpdateLastLogTime(q, i.Body.Lines[len(i.Body.Lines)-1].Time)
		if err != nil {
			return nil, err
		}
	}

	tm := huma_utils.TimeMeasure(c, "timeCreateManyBatches")
	err = querier.CreateManyBatches(q, lines, 100, false)
	if err != nil {
		return nil, err
	}
	tm()

	tm = huma_utils.TimeMeasure(c, "timeAddLogMetadataTotalBytes")
	err = dmodel.AddLogMetadataTotalBytes(q, lm.ID, addBytes)
	if err != nil {
		return nil, err
	}
	tm()

	return &huma_utils.Empty{}, nil
}

type restListLogsInput struct {
	OwnerType string `query:"owner_type"`
	OwnerId   string `query:"owner_id"`
}

func (s *LogsServer) restListLogs(c context.Context, i *restListLogsInput) (*huma_utils.List[models.LogMetadataModel], error) {
	q := querier.GetQuerier(c)
	w := auth_middleware.GetWorkspace(c)

	if i.OwnerType == "" || i.OwnerId == "" {
		return nil, huma.Error400BadRequest("missing owner_type/owner_id")
	}

	var machineId, boxId *string
	switch i.OwnerType {
	case "machine":
		machineId = &i.OwnerId
	case "box":
		boxId = &i.OwnerId
	}

	l, err := dmodel.ListLogMetadataForOwner(q, &w.ID, machineId, boxId, true)
	if err != nil {
		return nil, err
	}

	var ret []models.LogMetadataModel
	for _, lm := range l {
		x, err := models.LogMetadataFromDB(lm)
		if err != nil {
			return nil, err
		}
		ret = append(ret, *x)
	}

	return huma_utils.NewList(ret, len(ret)), nil
}

type sseLogsStreamInput struct {
	huma_utils.IdByPath

	LogId string `path:"logId"`
	Since string `query:"since"`
}

type endOfHistory struct {
}

func (s *LogsServer) sseLogsStream(c context.Context, i *sseLogsStreamInput, send sse.Sender) {
	err := s.sseLogsStreamErr(c, i, send)
	if err != nil {
		slog.ErrorContext(c, "error in sseLogsStreamErr", slog.Any("error", err))
		err = send.Data(models.LogsError{
			Message: err.Error(),
		})
		if err != nil {
			slog.ErrorContext(c, "error while sending sse error", slog.Any("error", err))
		}
	}
}

func (s *LogsServer) sseLogsStreamErr(c context.Context, i *sseLogsStreamInput, send sse.Sender) error {
	q := querier.GetQuerier(c)
	w := auth_middleware.GetWorkspace(c)

	lm, err := dmodel.GetLogMetadataById(q, &w.ID, i.LogId, true)
	if err != nil {
		return err
	}
	lm2, err := models.LogMetadataFromDB(*lm)
	if err != nil {
		return err
	}

	err = send.Data(lm2)
	if err != nil {
		return err
	}

	lastId := int64(-1)
	var since *time.Time

	if i.Since != "" {
		d, err := time.ParseDuration(i.Since)
		if err == nil {
			since = util.Ptr(time.Now().Add(-d))
		} else {
			j := "\"" + i.Since + "\""
			err = json.Unmarshal([]byte(j), &since)
			if err != nil {
				return huma.Error400BadRequest("invalid since argument")
			}
		}
	}

	didSendEndOfHistory := false
	for {
		var lines []dmodel.LogLine
		if since != nil {
			lines, err = dmodel.ListLogLinesSinceTime(q, lm.ID, *since, util.Ptr(int64(1000)))
		} else {
			lines, err = dmodel.ListLogLinesSinceSeq(q, lm.ID, lastId+1, util.Ptr(int64(1000)))
		}
		if err != nil {
			if !querier.IsSqlNotFoundError(err) {
				return err
			}
		}

		if len(lines) != 0 {
			// after the first line, go by sequence instead of time
			since = nil

			batchSize := 128
			for i := 0; i < len(lines) && c.Err() == nil; i += batchSize {
				e := min(i+batchSize, len(lines))
				b := lines[i:e]
				lb := boxspec.LogsBatch{
					Lines: make([]boxspec.LogsLine, 0, len(b)),
				}
				for _, l := range b {
					lm := models.LogLineFromDB(l)
					lb.Lines = append(lb.Lines, lm)
					lastId = l.ID
				}
				lb.Seq = lastId
				err = send.Data(lb)
				if err != nil {
					return err
				}
			}
		} else {
			if !didSendEndOfHistory {
				didSendEndOfHistory = true
				err = send.Data(endOfHistory{})
				if err != nil {
					return err
				}
			}

			// only sleep when no lines received
			select {
			case <-time.After(time.Second * 2):
				break
			case <-c.Done():
				return nil
			}
		}
	}
}
