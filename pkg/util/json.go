package util

import (
	"encoding/json"
	"os"
)

func UnmarshalJsonFile[T any](file string) (*T, error) {
	ret, _, err := UnmarshalJsonFileWithBytes[T](file)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func UnmarshalJsonFileWithBytes[T any](file string) (*T, []byte, error) {
	b, err := os.ReadFile(file)
	if err != nil {
		return nil, nil, err
	}
	var ret T
	err = json.Unmarshal(b, &ret)
	if err != nil {
		return nil, nil, err
	}
	return &ret, b, nil
}

func UnmarshalJsonFileWithHash[T any](file string) (*T, string, error) {
	ret, b, err := UnmarshalJsonFileWithBytes[T](file)
	if err != nil {
		return nil, "", err
	}
	return ret, Sha256Sum(b), nil
}
