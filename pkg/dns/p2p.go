package dns

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/koobox/unboxed/pkg/types"
	"github.com/koobox/unboxed/pkg/util"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-kad-dht"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	drouting "github.com/libp2p/go-libp2p/p2p/discovery/routing"
	dutil "github.com/libp2p/go-libp2p/p2p/discovery/util"
	"github.com/multiformats/go-multiaddr"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"
)

type DnsLibP2P struct {
	Store *DnsStore

	NetworkDomain  string
	BootstrapPeers []string

	dnsAnnouncement atomic.Pointer[types.DnsAnnouncement]

	host host.Host

	dht        *dht.IpfsDHT
	topic      *pubsub.Topic
	subscriber *pubsub.Subscription
}

func (d *DnsLibP2P) SetDnsAnnouncement(hostname string, ip string) {
	d.dnsAnnouncement.Store(&types.DnsAnnouncement{
		Hostname: hostname,
		IP:       ip,
	})
}

func (d *DnsLibP2P) Start(ctx context.Context) error {
	slog.InfoContext(ctx, "starting libp2p dns pub/sub")

	listenAddrs := []string{
		"/ip4/0.0.0.0/tcp/0",
	}
	h, err := libp2p.New(libp2p.ListenAddrStrings(listenAddrs...))
	if err != nil {
		return err
	}

	d.host = h

	slog.InfoContext(ctx, "libp2p host created", slog.Any("peerId", d.host.ID().String()), slog.Any("peerAddrs", d.host.Addrs()))

	err = d.createDHT(ctx)
	if err != nil {
		return err
	}

	slog.InfoContext(ctx, "creating gossip sub")
	gossipSub, err := pubsub.NewGossipSub(ctx, h)
	if err != nil {
		return err
	}

	topicName := fmt.Sprintf("_unboxed.%s", d.NetworkDomain)
	slog.InfoContext(ctx, "joining topic", slog.Any("topic", topicName))
	d.topic, err = gossipSub.Join(topicName)
	if err != nil {
		return err
	}

	slog.InfoContext(ctx, "subscribing topic")
	d.subscriber, err = d.topic.Subscribe()
	if err != nil {
		return err
	}

	go d.runPeerDiscovery(ctx)
	go d.runPublishDns(ctx)
	go d.runSubscribeDns(ctx)

	return nil
}

func (d *DnsLibP2P) createDHT(ctx context.Context) error {
	slog.InfoContext(ctx, "creating DHT")

	var bootstrapPeers []multiaddr.Multiaddr
	for _, p := range d.BootstrapPeers {
		pp, err := multiaddr.NewMultiaddr(p)
		if err != nil {
			return err
		}
		bootstrapPeers = append(bootstrapPeers, pp)
	}
	if len(bootstrapPeers) == 0 {
		bootstrapPeers = dht.DefaultBootstrapPeers
	}

	var bootstrapPeers2 []peer.AddrInfo
	for _, p := range bootstrapPeers {
		peerInfo, err := peer.AddrInfoFromP2pAddr(p)
		if err != nil {
			return err
		}
		bootstrapPeers2 = append(bootstrapPeers2, *peerInfo)
	}

	kdht, err := dht.New(ctx, d.host, dht.BootstrapPeers(bootstrapPeers2...))
	if err != nil {
		panic(err)
	}
	d.dht = kdht

	slog.InfoContext(ctx, "bootstrapping dht")
	if err = kdht.Bootstrap(ctx); err != nil {
		return err
	}
	<-kdht.RefreshRoutingTable()

	slog.InfoContext(ctx, "connecting to bootstrap peers")
	var wg sync.WaitGroup
	for _, peerInfo := range bootstrapPeers2 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			log := slog.With(slog.Any("peerId", peerInfo.ID), slog.Any("peerAddrs", peerInfo.Addrs))
			err := d.host.Connect(ctx, peerInfo)
			if err != nil {
				log.WarnContext(ctx, "failed to connect to bootstrap peer", slog.Any("error", err))
			} else {
				log.InfoContext(ctx, "connected to bootstrap peer")
			}
		}()
	}
	wg.Wait()

	return nil
}

func (d *DnsLibP2P) runPeerDiscovery(ctx context.Context) {
	discoveryNamespace := fmt.Sprintf("_unboxed.%s", d.NetworkDomain)

	slog.InfoContext(ctx, "announcing ourselves")
	routingDiscovery := drouting.NewRoutingDiscovery(d.dht)
	dutil.Advertise(ctx, routingDiscovery, discoveryNamespace)
	slog.InfoContext(ctx, "successfully announced ourself")

	util.LoopWithPrintErr(ctx, "runPeerDiscoveryOnce", 5*time.Second, func() error {
		return d.runPeerDiscoveryOnce(ctx, routingDiscovery, discoveryNamespace)
	})
}

func (d *DnsLibP2P) runPeerDiscoveryOnce(ctx context.Context, discovery *drouting.RoutingDiscovery, discoveryNamespace string) error {
	slog.InfoContext(ctx, "discovering peers...")
	peerChan, err := discovery.FindPeers(ctx, discoveryNamespace)
	if err != nil {
		return err
	}
	for peer := range peerChan {
		if peer.ID == d.host.ID() {
			continue
		}
		if len(peer.Addrs) == 0 {
			continue
		}

		log := slog.With(slog.Any("peerId", peer.ID), slog.Any("peerAddrs", peer.Addrs))
		log.InfoContext(ctx, "connecting to peer")
		err := d.host.Connect(ctx, peer)
		if err != nil {
			log.ErrorContext(ctx, "failed connecting to peer", slog.Any("error", err))
		} else {
			log.InfoContext(ctx, "connected to peer")
		}
	}

	slog.InfoContext(ctx, "runPeerDiscoveryOnce exited")
	return nil
}

func (d *DnsLibP2P) runPublishDns(ctx context.Context) {
	util.LoopWithPrintErr(ctx, "runPublishDns", 10*time.Second, func() error {
		return d.runPublishDnsOnce(ctx)
	})
}

func (d *DnsLibP2P) runPublishDnsOnce(ctx context.Context) error {
	a := d.dnsAnnouncement.Load()
	if a == nil {
		return nil
	}
	b, err := json.Marshal(a)
	if err != nil {
		return err
	}
	slog.InfoContext(ctx, "announcing dns info", slog.Any("dnsAnnouncement", *a))
	err = d.topic.Publish(ctx, b)
	if err != nil {
		return err
	}
	return nil
}

func (d *DnsLibP2P) runSubscribeDns(ctx context.Context) {
	util.LoopWithPrintErr(ctx, "runSubscribeDnsOnce", 5*time.Second, func() error {
		return d.runSubscribeDnsOnce(ctx)
	})
}

func (d *DnsLibP2P) runSubscribeDnsOnce(ctx context.Context) error {
	for {
		msg, err := d.subscriber.Next(ctx)
		if err != nil {
			return err
		}

		var dnsAnnouncement types.DnsAnnouncement
		err = json.Unmarshal(msg.Data, &dnsAnnouncement)
		if err != nil {
			return err
		}
		slog.InfoContext(ctx, "received dns announcement", slog.Any("msg", dnsAnnouncement), slog.Any("peerId", msg.ReceivedFrom))

		d.Store.Set(fmt.Sprintf("%s.%s.", dnsAnnouncement.Hostname, d.NetworkDomain), dnsAnnouncement.IP)
	}
}
