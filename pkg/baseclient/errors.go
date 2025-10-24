package baseclient

import (
	"errors"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
)

func IsNotFound(err error) bool {
	var err2 *huma.ErrorModel
	if errors.As(err, &err2) {
		if err2.Status == http.StatusNotFound {
			return true
		}
	}
	return false
}

func IsUnauthorized(err error) bool {
	var err2 *huma.ErrorModel
	if errors.As(err, &err2) {
		if err2.Status == http.StatusUnauthorized {
			return true
		}
	}
	return false
}
