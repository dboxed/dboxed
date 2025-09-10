package dns_proxy

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"net"
	"os/exec"
	"reflect"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/dboxed/dboxed-common/util"
	util2 "github.com/dboxed/dboxed/pkg/util"
	"github.com/miekg/dns"
	"github.com/vishvananda/netns"
)

type DnsProxy struct {
	ListenNamespace netns.NsHandle
	QueryNamespace  netns.NsHandle
	ListenIP        net.IP

	HostFsPath string

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
		go func() {
			err = d.runRequestsThread(ctx)
			if err != nil {
				slog.ErrorContext(ctx, "error while running host thread", slog.Any("error", err))
			}
		}()
	}

	go func() {
		util.LoopWithPrintErr(ctx, "readHostResolvConf", 5*time.Second, func() error {
			return d.readHostResolvConf(ctx)
		})
	}()

	d.startServing(ctx, d.udpServer)
	d.startServing(ctx, d.tcpServer)

	return nil
}

func (d *DnsProxy) readHostResolvConf(ctx context.Context) error {
	// we need to enter the hostfs with chroot as otherwise links won't resolve properly
	// using chroot+cat is the simplest way here, but we might need to consider using unshare in some way
	buf := bytes.NewBuffer(nil)
	cmd := exec.CommandContext(ctx, "chroot", d.HostFsPath, "cat", "/etc/resolv.conf")
	cmd.Stdout = buf
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to read host resolv.conf: %w", err)
	}

	c, err := dns.ClientConfigFromReader(buf)
	if err != nil {
		return err
	}
	old := d.resolveConf.Swap(c)

	if !reflect.DeepEqual(old, c) {
		slog.InfoContext(ctx, "using host nameservers", slog.Any("nameservers", d.resolveConf.Load().Servers), slog.Any("port", d.resolveConf.Load().Port))
	}
	return nil
}

func (d *DnsProxy) runRequestsThread(ctx context.Context) error {
	return util2.RunInNetNs(d.QueryNamespace, func() error {
		return d.runRequestsThreadInNs(ctx)
	})
}

func (d *DnsProxy) runRequestsThreadInNs(ctx context.Context) error {
	for r := range d.requests {
		d.handleRequest(ctx, r)
	}
	return nil
}

func (d *DnsProxy) handleRequest(ctx context.Context, r dnsRequest) {
	resolveConf := d.resolveConf.Load()
	dnsResolver := net.JoinHostPort(resolveConf.Servers[0], resolveConf.Port)

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

	log.InfoContext(ctx, "start listen")
	err := util2.RunInNetNs(d.ListenNamespace, func() error {
		var err error
		if isTcp {
			dnsServer.Listener, err = net.Listen(dnsNet, dnsServer.Addr)
		} else {
			dnsServer.PacketConn, err = net.ListenPacket(dnsNet, dnsServer.Addr)
		}
		return err
	})
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
