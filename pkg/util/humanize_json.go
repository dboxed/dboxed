package util

import (
	"encoding/json"

	"github.com/dustin/go-humanize"
)

type HumanBytes struct {
	Bytes int64
}

func (v *HumanBytes) UnmarshalJSON(b []byte) error {
	var str string
	err := json.Unmarshal(b, &str)
	if err != nil {
		return err
	}

	pd, err := humanize.ParseBytes(str)
	if err != nil {
		return err
	}
	v.Bytes = int64(pd)
	return nil
}

func (v *HumanBytes) MarshalJSON() ([]byte, error) {
	return json.Marshal(humanize.Bytes(uint64(v.Bytes)))
}
