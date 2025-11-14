package load_balancer

type LoadBalancerCommands struct {
	Create CreateCmd `cmd:"" help:"Create an load balancer"`
	Get    GetCmd    `cmd:"" help:"Get an load balancer"`
	List   ListCmd   `cmd:"" help:"List load balancers" aliases:"ls"`
	Update UpdateCmd `cmd:"" help:"Update an load balancer"`
	Delete DeleteCmd `cmd:"" help:"Delete an load balancer" aliases:"rm,delete"`
}
