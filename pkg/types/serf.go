package types

type SerfMembers struct {
	Members []SerfNode `json:"members"`
}

type SerfNode struct {
	Name string `json:"name"`
	Addr string `json:"addr"`
}
