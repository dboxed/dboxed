package utils

import "fmt"

const DefaultWorkdir = "/var/lib/unboxed"

func GetBoxUrl(urlArg string, fileArg string) (string, error) {
	if urlArg == "" && fileArg == "" {
		return "", fmt.Errorf("either --box-url or --box-file must be set")
	} else if urlArg != "" && fileArg != "" {
		return "", fmt.Errorf("only one of --box-url or --box-file must be set")
	}

	url := urlArg
	if url == "" {
		url = "file://" + fileArg
	}
	return url, nil
}
