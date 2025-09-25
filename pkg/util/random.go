package util

import "k8s.io/apimachinery/pkg/util/rand"

func RandomString(length int) string {
	return rand.String(length)
}
