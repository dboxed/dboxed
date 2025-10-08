package boxes

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/sse"
	"github.com/dboxed/dboxed/pkg/boxspec"
	"github.com/dboxed/dboxed/pkg/server/db/dmodel"
	"github.com/dboxed/dboxed/pkg/server/db/querier"
	"github.com/dboxed/dboxed/pkg/server/global"
	"github.com/dboxed/dboxed/pkg/server/huma_utils"
	"github.com/dboxed/dboxed/pkg/server/models"
	"github.com/dboxed/dboxed/pkg/util"
)

func (s *BoxesServer) putLogMetadata(c context.Context, boxId int64, logMetadata boxspec.LogMetadata) (*dmodel.LogMetadata, error) {
	q := querier.GetQuerier(c)
	w := global.GetWorkspace(c)

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
		BoxID:    boxId,
		FileName: logMetadata.FileName,
		Format:   logMetadata.Format,
		Metadata: string(metadataBytes),
	}
	err := lm.CreateOrUpdate(q)
	if err != nil {
		return nil, err
	}
	return &lm, nil
}

func (s *BoxesServer) restPostLogs(c context.Context, i *huma_utils.IdByPathAndJsonBody[models.PostLogs]) (*huma_utils.Empty, error) {
	q := querier.GetQuerier(c)

	dm, err := s.putLogMetadata(c, i.Id, i.Body.Metadata)
	if err != nil {
		return nil, err
	}

	lines := make([]*dmodel.LogLine, 0, len(i.Body.Lines))

	for _, ll := range i.Body.Lines {
		dl := &dmodel.LogLine{
			LogID: dm.ID,
			Line:  ll.Line,
			Time:  ll.Time,
		}
		lines = append(lines, dl)
	}

	logFileName := i.Body.Metadata.FileName
	if len(logFileName) > 20 {
		logFileName = logFileName[:3] + "..." + logFileName[len(logFileName)-14:]
	}
	huma_utils.ExtraLogValue(c, "lineCount", len(lines))
	huma_utils.ExtraLogValue(c, "logId", dm.ID)
	huma_utils.ExtraLogValue(c, "fileName", logFileName)

	err = querier.CreateManyBatches(q, lines, 100, false)
	if err != nil {
		return nil, err
	}

	return &huma_utils.Empty{}, nil
}

func (s *BoxesServer) restListLogs(c context.Context, i *huma_utils.IdByPath) (*huma_utils.List[models.LogMetadataModel], error) {
	q := querier.GetQuerier(c)
	w := global.GetWorkspace(c)

	l, err := dmodel.ListLogMetadataForBox(q, &w.ID, i.Id, true)
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

	LogId int64  `path:"logId"`
	Since string `query:"since"`
}

type endOfHistory struct {
}

func (s *BoxesServer) sseLogsStream(c context.Context, i *sseLogsStreamInput, send sse.Sender) {
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

func (s *BoxesServer) sseLogsStreamErr(c context.Context, i *sseLogsStreamInput, send sse.Sender) error {
	q := querier.GetQuerier(c)
	w := global.GetWorkspace(c)

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
			lines, err = dmodel.ListLogLinesSinceTime(q, lm.ID, *since)
		} else {
			lines, err = dmodel.ListLogLinesSinceSeq(q, lm.ID, lastId+1)
		}
		if err != nil {
			if !querier.IsSqlNotFoundError(err) {
				return err
			}
		}
		for _, l := range lines {
			lm := models.LogLineFromDB(l)
			err = send.Data(lm)
			if err != nil {
				return err
			}
			lastId = l.ID
		}

		if len(lines) != 0 {
			// after the first line, go by sequence instead of time
			since = nil
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
