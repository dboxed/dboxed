package ingress_proxy

type IngressProxyCommands struct {
	Create CreateCmd `cmd:"" help:"Create an ingress proxy"`
	Get    GetCmd    `cmd:"" help:"Get an ingress proxy"`
	List   ListCmd   `cmd:"" help:"List ingress proxies" aliases:"ls"`
	Update UpdateCmd `cmd:"" help:"Update an ingress proxy"`
	Delete DeleteCmd `cmd:"" help:"Delete an ingress proxy" aliases:"rm,delete"`
}
