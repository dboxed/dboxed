package models

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"time"

	"github.com/mr-tron/base58"
	"github.com/nats-io/nkeys"
)

const CurrentTokenVersion = 1

type OldToken struct {
	TokenVersion int64
	WorkspaceId  int64
	BoxId        int64
	Exp          int64
}

func (t *OldToken) serialize() ([]byte, error) {
	if t.TokenVersion != CurrentTokenVersion {
		return nil, fmt.Errorf("can only serialize current token version")
	}

	tokenBytes := make([]byte, 0, 128) // also include some space for the sig
	tokenBytes = binary.AppendVarint(tokenBytes, CurrentTokenVersion)
	tokenBytes = binary.AppendVarint(tokenBytes, t.WorkspaceId)
	tokenBytes = binary.AppendVarint(tokenBytes, t.BoxId)
	tokenBytes = binary.AppendVarint(tokenBytes, t.Exp)

	return tokenBytes, nil
}

func (t *OldToken) deserialize(b []byte) ([]byte, error) {
	r := bytes.NewReader(b)
	version, err := binary.ReadVarint(r)
	if err != nil {
		return nil, err
	}
	if version != 1 {
		return nil, fmt.Errorf("only version 1 is supported")
	}

	t.TokenVersion = version
	t.WorkspaceId, err = binary.ReadVarint(r)
	if err != nil {
		return nil, err
	}
	t.BoxId, err = binary.ReadVarint(r)
	if err != nil {
		return nil, err
	}
	t.Exp, err = binary.ReadVarint(r)
	if err != nil {
		return nil, err
	}

	offset := int(r.Size()) - r.Len()
	sig := b[offset:]
	return sig, nil
}

func (t *OldToken) BuildTokenStr(nkeySeed []byte) (string, error) {
	nkeyPair, err := nkeys.FromSeed(nkeySeed)
	if err != nil {
		return "", err
	}

	b, err := t.serialize()
	if err != nil {
		return "", err
	}

	sig, err := nkeyPair.Sign(b)
	if err != nil {
		return "", err
	}
	b = append(b, sig...)
	tokenStr := base58.Encode(b)

	return tokenStr, nil
}

func (t *OldToken) Verify(nkey string, sig []byte) error {
	exp := time.Unix(t.Exp, 0)
	if time.Now().After(exp) {
		return fmt.Errorf("token is expired")
	}

	pk, err := nkeys.FromPublicKey(nkey)
	if err != nil {
		return err
	}

	b, err := t.serialize()
	if err != nil {
		return err
	}

	return pk.Verify(b, sig)
}

func TokenFromStr(tokenStr string) (*OldToken, []byte, error) {
	tokenBytes, err := base58.Decode(tokenStr)
	if err != nil {
		return nil, nil, err
	}
	var t OldToken
	sig, err := t.deserialize(tokenBytes)
	if err != nil {
		return nil, nil, err
	}

	return &t, sig, nil
}
