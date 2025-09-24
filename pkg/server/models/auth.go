package models

type AuthInfo struct {
	OidcIssuerUrl string `json:"oidcIssuerUrl"`
	OidcClientId  string `json:"oidcClientId"`
}
