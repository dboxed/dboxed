package dns

type DnsPubSub interface {
	SetDnsAnnouncement(hostname string, ip string)
}
