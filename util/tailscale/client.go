package tailscale

// This code is used to pull machine names from Tailscale and insert them as DNS records in a zone.
// Heavily based on https://github.com/damomurf/coredns-tailscale/blob/4b04cdf112741de900bf26effd77dcb27a4b0d4a/tailscale.go

import (
	"context"
	"github.com/henrikvtcodes/tungsten/util"
	"net"
	"strings"
	"sync"
	tsLocal "tailscale.com/client/local"
	"tailscale.com/ipn"
	"tailscale.com/tailcfg"
	"tailscale.com/tsnet"
	"tailscale.com/types/netmap"
	"time"
)

type MachineEntry struct {
	Name        string
	ARecords    []net.IP
	AAAARecords []net.IP
}

type CNameEntry struct {
	Name    string
	CNameTo []string
}

type Tailscale struct {
	zone string

	authkey  string
	hostname string
	srv      *tsnet.Server
	lc       *tsLocal.Client

	mu             sync.RWMutex
	MachineEntries map[string]MachineEntry
	CNameEntries   map[string]CNameEntry
}

// Start connects the Tailscale plugin to a tailscale daemon and populates DNS Entries for nodes in the tailnet.
// DNS Entries are automatically kept up to date with any node changes.
//
// If t.authkey is non-empty, this function uses that key to connect to the Tailnet using a tsnet server
// instead of connecting to the local tailscaled instance.
func (t *Tailscale) Start() error {
	if t.authkey != "" {
		hostname := t.hostname
		if t.hostname == "" {
			hostname = "coredns"
		}
		// authkey was provided, so startup a local tsnet server
		t.srv = &tsnet.Server{
			Hostname:     hostname,
			AuthKey:      t.authkey,
			Logf:         util.Logger.Debug().Msgf,
			RunWebClient: true,
		}
		err := t.srv.Start()
		if err != nil {
			return err
		}
		t.lc, err = t.srv.LocalClient()
		if err != nil {
			return err
		}
	} else {
		// zero value LocalClient will connect to local tailscaled
		t.lc = &tsLocal.Client{}
	}

	util.Logger.Debug().Msg("TS Client Run: Watching IPN Bus")
	go t.watchIPNBus()
	return nil
}

// watchIPNBus watches the Tailscale IPN Bus and updates DNS Entries for any netmap update.
// This function does not return. If it is unable to read from the IPN Bus, it will continue to retry.
func (t *Tailscale) watchIPNBus() {
	for {
		watcher, err := t.lc.WatchIPNBus(context.Background(), ipn.NotifyInitialNetMap|ipn.NotifyNoPrivateKeys)
		if err != nil {
			util.Logger.Info().Msg("Tailscale IPN: Unable to read from Tailscale event bus, retrying in 1 minute")
			time.Sleep(1 * time.Minute)
			continue
		}
		defer watcher.Close()

		for {
			n, err := watcher.Next()
			if err != nil {
				// If we're unable to read, then close watcher and reconnect
				_ = watcher.Close()
				break
			}
			t.processNetMap(n.NetMap)
		}
	}
}

func (t *Tailscale) processNetMap(nm *netmap.NetworkMap) {
	if nm == nil {
		return
	}

	util.Logger.Debug().Msgf("Self tags: %+v", nm.SelfNode.Tags().AsSlice())
	nodes := []tailcfg.NodeView{nm.SelfNode}
	nodes = append(nodes, nm.Peers...)

	entries := map[string]map[string][]string{}
	machineEntries := map[string]MachineEntry{}
	cnameEntries := map[string]CNameEntry{}
	for _, node := range nodes {
		if node.IsWireGuardOnly() {
			// IsWireGuardOnly identifies a node as a Mullvad exit node.
			continue
		}
		if !node.Sharer().IsZero() {
			// Skip shared nodes, since they don't necessarily have unique hostnames within this tailnet.
			// TODO: possibly make it configurable to include shared nodes and figure out what hostname to use.
			continue
		}

		hostname := node.ComputedName()
		mEntry, mOk := machineEntries[hostname]

		if !mOk {
			mEntry = MachineEntry{}
		}

		for _, pfx := range node.Addresses().AsSlice() {
			addr := pfx.Addr()
			if addr.Is4() {
				mEntry.ARecords = append(mEntry.ARecords, net.ParseIP(addr.String()))
			} else if addr.Is6() {
				mEntry.AAAARecords = append(mEntry.AAAARecords, net.ParseIP(addr.String()))
			}
		}

		// Process Tags looking for cname- prefixed ones
		if node.Tags().Len() > 0 {
			for _, raw := range node.Tags().AsSlice() {
				if tag, ok := strings.CutPrefix(raw, "tag:cname-"); ok {
					if _, ok := cnameEntries[tag]; !ok {
						cnameEntries[tag] = CNameEntry{}
					}

					cnameEntries[tag] = CNameEntry{Name: tag, CNameTo: append(cnameEntries[tag].CNameTo, hostname)}
				}
			}
		}

		machineEntries[hostname] = mEntry
	}

	t.mu.Lock()
	t.MachineEntries = machineEntries
	t.CNameEntries = cnameEntries
	t.mu.Unlock()
	util.Logger.Debug().Msgf("Updated %d Tailscale Entries", len(entries))
}

func (t *Tailscale) FindMachine(hostname string) (*MachineEntry, bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	for _, m := range t.MachineEntries {
		if m.Name == hostname {
			return &m, true
		}
	}
	return nil, false
}

func (t *Tailscale) FindCNameEntry(subdomain string) (*CNameEntry, bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	for _, c := range t.CNameEntries {
		if subdomain == c.Name {
			return &c, true
		}
	}
	return nil, false
}
