package util

import (
	"encoding/json"
	"os"
)

func ReadJsonFile[T any](p string) (*T, error) {
	b, err := os.ReadFile(p)
	if err != nil {
		return nil, err
	}

	var x T
	err = json.Unmarshal(b, &x)
	if err != nil {
		return nil, err
	}
	return &x, nil
}
