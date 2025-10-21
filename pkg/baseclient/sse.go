package baseclient

import (
	"bufio"
	"context"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type SSEMessage struct {
	Event string
	ID    int
	Data  string
	Retry int
}

func RequestApiSSE(ctx context.Context, c *Client, p string, q url.Values, cb func(m SSEMessage) error) error {
	header := http.Header{}
	header.Set("Accept", "text/event-stream")

	resp, err := requestApiResponse(ctx, c, "GET", p, q, struct{}{}, true, &header)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	reader := bufio.NewReader(resp.Body)

	for {
		var dataLines []string

		var msg SSEMessage
		firstLine := true
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				if err == io.EOF {
					if firstLine {
						return io.ErrUnexpectedEOF
					}
					return nil
				}
				return err
			}
			firstLine = false

			if len(line) > 0 && line[len(line)-1] == '\n' {
				line = line[:len(line)-1]
			}
			if line == "" {
				msg.Data = strings.Join(dataLines, "\n")
				break
			}
			if strings.HasPrefix(line, "event: ") {
				msg.Event = strings.TrimPrefix(line, "event: ")
			} else if strings.HasPrefix(line, "id: ") {
				id, err := strconv.ParseInt(strings.TrimPrefix(line, "id: "), 10, 32)
				if err != nil {
					return err
				}
				msg.ID = int(id)
			} else if strings.HasPrefix(line, "data: ") {
				dataLines = append(dataLines, strings.TrimPrefix(line, "data: "))
			}
		}

		err = cb(msg)
		if err != nil {
			return err
		}
	}
}
