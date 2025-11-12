package ingress_proxy

type IngressProxyCommands struct {
	Create CreateCmd `cmd:"" help:"Create an ingress proxy"`
	Get    GetCmd    `cmd:"" help:"Get an ingress proxy"`
	List   ListCmd   `cmd:"" help:"List ingress proxies" aliases:"ls"`
	Delete DeleteCmd `cmd:"" help:"Delete an ingress proxy" aliases:"rm,delete"`
}
