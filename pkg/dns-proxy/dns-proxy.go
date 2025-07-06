package dns_proxy

import (
	"context"
	"fmt"
	"github.com/koobox/unboxed/pkg/util"
	"github.com/miekg/dns"
	"github.com/vishvananda/netns"
	"log/slog"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type DnsProxy struct {
	ListenNamespace netns.NsHandle
	QueryNamespace  netns.NsHandle
	ListenIP        net.IP

	staticHostsMapMutex sync.Mutex
	staticHostsMap      map[string]string

	udpServer *dns.Server
	tcpServer *dns.Server

	resolveConf *dns.ClientConfig
	udpClient   *dns.Client
	tcpClient   *dns.Client

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

	var err error
	d.resolveConf, err = dns.ClientConfigFromFile("/etc/resolv.conf")
	if err != nil {
		return err
	}

	slog.InfoContext(ctx, "starting dns proxy")
	slog.InfoContext(ctx, "using host nameservers", slog.Any("nameservers", d.resolveConf.Servers), slog.Any("port", d.resolveConf.Port))

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
		go func() {
			err = d.runRequestsThread(ctx)
			if err != nil {
				slog.ErrorContext(ctx, "error while running host thread", slog.Any("error", err))
			}
		}()
	}

	return nil
}

func (d *DnsProxy) WriteResolvConf(root string) error {
	resolveConf := fmt.Sprintf(`# This is the unboxed dns proxy, which listens inside the sandboxed network namespace
# and forwards requests to the host's resolv.conf nameservers
nameserver %s
search .
`, d.ListenIP.String())

	err := os.WriteFile(filepath.Join(root, "etc/resolv.conf"), []byte(resolveConf), 0666)
	if err != nil {
		return err
	}
	return nil
}

func (d *DnsProxy) runRequestsThread(ctx context.Context) error {
	return util.RunInNetNs(d.QueryNamespace, func() error {
		return d.runRequestsThreadInNs(ctx)
	})
}

func (d *DnsProxy) runRequestsThreadInNs(ctx context.Context) error {
	dnsResolver := net.JoinHostPort(d.resolveConf.Servers[0], d.resolveConf.Port)

	for r := range d.requests {
		d.handleRequest(ctx, dnsResolver, r)
	}
	return nil
}

func (d *DnsProxy) handleRequest(ctx context.Context, dnsResolver string, r dnsRequest) {
	log := slog.With(slog.Any("id", r.request.Id), slog.Any("tcp", r.tcp), slog.Any("dnsResolver", dnsResolver))
	//log.DebugContext(ctx, "handling request"+r.request.String())

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
	//log.DebugContext(ctx, "responding with "+resp.String())
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
		//slog.DebugContext(ctx, "queuing request"+r.String())
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

	log.InfoContext(ctx, "start listen and serve for dns server")

	go func() {
		err := util.RunInNetNs(d.ListenNamespace, func() error {
			return dnsServer.ListenAndServe()
		})
		if err != nil {
			slog.ErrorContext(ctx, "error while serving dns server")
		}
	}()

	return dnsServer, nil
}
