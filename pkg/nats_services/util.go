package nats_services

import (
	"os"

	"github.com/nats-io/nkeys"
)

func loadSeed(p string) (nkeys.KeyPair, string, error) {
	seed, err := os.ReadFile(p)
	if err != nil {
		return nil, "", err
	}
	kp, err := nkeys.FromSeed(seed)
	if err != nil {
		return nil, "", err
	}
	nkey, err := kp.PublicKey()
	if err != nil {
		return nil, "", err
	}
	return kp, nkey, nil
}
