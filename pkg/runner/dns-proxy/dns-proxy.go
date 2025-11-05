package dns_proxy

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"reflect"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/dboxed/dboxed/pkg/util"
	"github.com/miekg/dns"
)

type DnsProxy struct {
	ListenIP net.IP

	HostResolvConfFile string

	staticHostsMapMutex sync.Mutex
	staticHostsMap      map[string]string

	udpServer *dns.Server
	tcpServer *dns.Server

	resolveConf atomic.Pointer[dns.ClientConfig]

	udpClient *dns.Client
	tcpClient *dns.Client

	requests chan dnsRequest
}

type dnsRequest struct {
	tcp     bool
	writer  dns.ResponseWriter
	request *dns.Msg
}

func (d *DnsProxy) Start(ctx context.Context) error {
	d.staticHostsMap = map[string]string{}
	d.requests = make(chan dnsRequest)

	err := d.readHostResolvConf(ctx)
	if err != nil {
		return err
	}

	if _, err = d.getResolver(); err != nil {
		return err
	}

	slog.InfoContext(ctx, "starting dns proxy")

	d.udpClient = &dns.Client{
		Net: "udp",
	}
	d.tcpClient = &dns.Client{
		Net: "tcp",
	}

	d.udpServer, err = d.startListen(ctx, "udp")
	if err != nil {
		return err
	}
	d.tcpServer, err = d.startListen(ctx, "tcp")
	if err != nil {
		return err
	}

	for range 8 {
		go d.runRequestsThread(ctx)
	}

	go func() {
		util.LoopWithPrintErr(ctx, "readHostResolvConf", 5*time.Second, func() error {
			return d.readHostResolvConf(ctx)
		})
	}()
	for d.resolveConf.Load() == nil {
		if !util.SleepWithContext(ctx, time.Millisecond*10) {
			return ctx.Err()
		}
	}

	d.startServing(ctx, d.udpServer)
	d.startServing(ctx, d.tcpServer)

	return nil
}

func (d *DnsProxy) readHostResolvConf(ctx context.Context) error {
	c, err := dns.ClientConfigFromFile(d.HostResolvConfFile)
	if err != nil {
		return err
	}
	old := d.resolveConf.Swap(c)

	if !reflect.DeepEqual(old, c) {
		slog.InfoContext(ctx, "using host nameservers", slog.Any("nameservers", d.resolveConf.Load().Servers), slog.Any("port", d.resolveConf.Load().Port))
	}
	return nil
}

func (d *DnsProxy) runRequestsThread(ctx context.Context) {
	for r := range d.requests {
		d.handleRequest(ctx, r)
	}
}

func (d *DnsProxy) getResolver() (string, error) {
	resolveConf := d.resolveConf.Load()
	if len(resolveConf.Servers) == 0 {
		return "", fmt.Errorf("nameservers missing in host resolv.conf")
	}

	var dnsResolver string
	foundIpv6 := false

	// find first ipv6 server
	for _, s := range resolveConf.Servers {
		ip := net.ParseIP(s)
		if ip == nil {
			continue
		}
		ip4 := ip.To4()
		if ip4 != nil {
			dnsResolver = net.JoinHostPort(ip4.String(), resolveConf.Port)
			break
		} else if ip.To16() != nil {
			foundIpv6 = true
		}
	}
	if dnsResolver == "" {
		if foundIpv6 {
			return "", fmt.Errorf("only found ipv6 nameservers in host resolv.conf, but only ipv4 nameservers are supported at the moment")
		} else {
			return "", fmt.Errorf("no ipv4 nameserver found in host resolv.conf")
		}
	}
	return dnsResolver, nil
}

func (d *DnsProxy) handleRequest(ctx context.Context, r dnsRequest) {
	dnsResolver, err := d.getResolver()
	if err != nil {
		slog.ErrorContext(ctx, err.Error())
		return
	}

	log := slog.With(slog.Any("id", r.request.Id), slog.Any("tcp", r.tcp), slog.Any("dnsResolver", dnsResolver))
	log.DebugContext(ctx, "handling request"+r.request.String())

	if len(r.request.Question) != 1 {
		m := &dns.Msg{}
		m.SetRcode(r.request, dns.RcodeNotImplemented)
		_ = r.writer.WriteMsg(m)
		return
	}

	staticResponse := d.resolveStaticHost(r.request.Question[0].Name)
	if staticResponse != nil {
		staticResponse.SetRcode(r.request, dns.RcodeSuccess)
		_ = r.writer.WriteMsg(staticResponse)
		return
	}

	client := d.udpClient
	if r.tcp {
		client = d.tcpClient
	}
	resp, rtt, err := client.Exchange(r.request, dnsResolver)
	if rtt != 0 {
		log = log.With(slog.Any("rtt", rtt))
	}
	if err != nil {
		log.Error("error while handling request", slog.Any("error", err))
		m := &dns.Msg{}
		m.SetRcode(r.request, dns.RcodeServerFailure)
		_ = r.writer.WriteMsg(m)
		return
	}
	log.DebugContext(ctx, "responding with "+resp.String())
	resp.SetReply(r.request)
	err = r.writer.WriteMsg(resp)
	if err != nil {
		log.Error("error while sending response", slog.Any("error", err))
		return
	}
}

func (d *DnsProxy) SetStaticHostsMap(m map[string]string) {
	d.staticHostsMapMutex.Lock()
	defer d.staticHostsMapMutex.Unlock()

	d.staticHostsMap = m
}

func (d *DnsProxy) resolveStaticHost(name string) *dns.Msg {
	d.staticHostsMapMutex.Lock()
	defer d.staticHostsMapMutex.Unlock()

	ip := d.staticHostsMap[name]
	if ip == "" {
		return nil
	}

	aRec := &dns.A{
		Hdr: dns.RR_Header{
			Name:   name,
			Rrtype: dns.TypeA,
			Class:  dns.ClassINET,
			Ttl:    10,
		},
		A: net.ParseIP(ip).To4(),
	}

	var msg dns.Msg
	msg.Compress = true
	msg.Answer = append(msg.Answer, aRec)

	return &msg
}

func (d *DnsProxy) startListen(ctx context.Context, dnsNet string) (*dns.Server, error) {
	log := slog.With(slog.Any("dnsNet", dnsNet))

	isTcp := strings.HasPrefix(dnsNet, "tcp")

	mux := dns.NewServeMux()
	mux.HandleFunc(".", func(writer dns.ResponseWriter, r *dns.Msg) {
		slog.DebugContext(ctx, "queuing request"+r.String())
		req := dnsRequest{
			tcp:     isTcp,
			writer:  writer,
			request: r,
		}
		d.requests <- req
	})

	dnsServer := &dns.Server{
		Addr:    net.JoinHostPort(d.ListenIP.String(), "53"),
		Net:     dnsNet,
		Handler: mux,
	}

	log.InfoContext(ctx, "start listen")

	var err error
	if isTcp {
		dnsServer.Listener, err = net.Listen(dnsNet, dnsServer.Addr)
	} else {
		dnsServer.PacketConn, err = net.ListenPacket(dnsNet, dnsServer.Addr)
	}
	if err != nil {
		return nil, err
	}

	return dnsServer, nil
}

func (d *DnsProxy) startServing(ctx context.Context, dnsServer *dns.Server) {
	go func() {
		err := dnsServer.ActivateAndServe()
		if err != nil {
			slog.ErrorContext(ctx, "error while serving dns server", slog.Any("error", err))
		}
	}()
}
