package util

import (
	"os"

	"sigs.k8s.io/yaml"
)

func UnmarshalYamlFile[T any](file string) (*T, error) {
	ret, _, err := UnmarshalYamlFileWithBytes[T](file)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func UnmarshalYamlFileWithBytes[T any](file string) (*T, []byte, error) {
	b, err := os.ReadFile(file)
	if err != nil {
		return nil, nil, err
	}
	var ret T
	err = yaml.Unmarshal(b, &ret)
	if err != nil {
		return nil, nil, err
	}
	return &ret, b, nil
}

func UnmarshalYamlFileWithHash[T any](file string) (*T, string, error) {
	ret, b, err := UnmarshalYamlFileWithBytes[T](file)
	if err != nil {
		return nil, "", err
	}
	return ret, Sha256Sum(b), nil
}
