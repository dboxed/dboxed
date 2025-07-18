package dns

import (
	"sync"
	"time"
)

type DnsStore struct {
	expiry  time.Duration
	entries map[string]dnsStoreEntry

	m sync.Mutex
}

type dnsStoreEntry struct {
	hostname string
	ip       string

	expiresAt time.Time
}

func NewDnsStore(expiry time.Duration) *DnsStore {
	return &DnsStore{
		expiry:  expiry,
		entries: map[string]dnsStoreEntry{},
	}
}

func (s *DnsStore) Set(hostname string, ip string) {
	s.m.Lock()
	defer s.m.Unlock()

	s.entries[hostname] = dnsStoreEntry{
		hostname:  hostname,
		ip:        ip,
		expiresAt: time.Now().Add(s.expiry),
	}
}

func (s *DnsStore) Map() map[string]string {
	s.m.Lock()
	defer s.m.Unlock()

	ret := map[string]string{}

	now := time.Now()
	for k, v := range s.entries {
		if now.After(v.expiresAt) {
			delete(s.entries, k)
			continue
		}
		ret[k] = v.ip
	}
	return ret
}
