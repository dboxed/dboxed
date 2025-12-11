package baseclient

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/dboxed/dboxed/pkg/server/models"
	"github.com/dboxed/dboxed/pkg/util"
	"golang.org/x/oauth2"
	"sigs.k8s.io/yaml"
)

type ClientAuth struct {
	ApiUrl      string           `json:"apiUrl"`
	AuthInfo    *models.AuthInfo `json:"authInfo"`
	WorkspaceId *string          `json:"workspaceId"`

	Oauth2Token *oauth2.Token `json:"oauth2Token"`
	StaticToken *string       `json:"staticToken"`
}

func GetDefaultClientAuthFile() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", nil
	}
	p := filepath.Join(homeDir, ".dboxed", "client-auth.yaml")
	return p, nil
}

func ReadClientAuth(clientAuthFile *string) (*ClientAuth, error) {
	if clientAuthFile == nil {
		p, err := GetDefaultClientAuthFile()
		if err != nil {
			return nil, err
		}
		if p == "" {
			return nil, os.ErrNotExist
		}
		clientAuthFile = &p
	}

	b, err := os.ReadFile(*clientAuthFile)
	if err != nil {
		return nil, err
	}

	var ret ClientAuth
	err = yaml.Unmarshal(b, &ret)
	if err != nil {
		return nil, err
	}
	return &ret, nil
}

func WriteClientAuth(clientAuthFile *string, ca *ClientAuth) error {
	if clientAuthFile == nil {
		p, err := GetDefaultClientAuthFile()
		if err != nil {
			return err
		}
		if p == "" {
			return fmt.Errorf("could not determine client auth path config path")
		}
		clientAuthFile = &p
	}
	err := os.MkdirAll(filepath.Dir(*clientAuthFile), 0700)
	if err != nil {
		return err
	}
	b, err := yaml.Marshal(ca)
	if err != nil {
		return err
	}

	err = util.AtomicWriteFile(*clientAuthFile, b, 0600)
	if err != nil {
		return err
	}
	return nil
}
