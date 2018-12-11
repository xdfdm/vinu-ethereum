package lachesis

import (
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"

	"github.com/ethereum/go-ethereum/event"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/discv5"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/ethereum/go-ethereum/rlp"
)

const (
	peerCount = 1 // see eth.minDesiredPeerCount = 5
)

var (
	// available protocol
	caps = []p2p.Cap{
		{Name: "eth", Version: 63},
	}
)

/*
 * Wrapper
 */

func NewServer(cfg p2p.Config) *p2p.Server {
	return p2p.WrapServer(&lachesisServer{Config: cfg})
}

/*
 * p2p.ServerInterface implementation
 */

// lachesisServer simulates other peers by lachesis.
type lachesisServer struct {
	// Config fields may not be modified while the server is running.
	p2p.Config

	lock    sync.Mutex // protects running
	running bool
	quit    chan struct{}
	wg      sync.WaitGroup

	peers    []*p2p.Peer
	peerFeed event.Feed

	log log.Logger
}

// Start starts running the server.
// Servers can not be re-used after stopping.
func (srv *lachesisServer) Start() (err error) {
	srv.lock.Lock()
	defer srv.lock.Unlock()
	if srv.running {
		return errors.New("server already running")
	}

	srv.log = srv.Config.Logger
	if srv.log == nil {
		srv.log = log.New()
	}

	err = srv.Config.LachesisAdapter.Start(srv.log)
	if err != nil {
		return
	}

	srv.running = true

	// make fake peers
	// peers should be sorted alphabetically by node identifier
	// (or sort it when PeersInfo())
	for i := 0; i < peerCount; i++ {
		id := enode.HexID(fmt.Sprintf("%#064x", i))
		name := fmt.Sprintf("fake-node-%d", i)
		peer := p2p.NewPeer(id, name, caps)
		srv.log.Debug("Fake peer created", "name", peer.Name(), "id", peer.ID())
		srv.peers = append(srv.peers, peer)
		// broadcast peer add
		srv.peerFeed.Send(&p2p.PeerEvent{
			Type: p2p.PeerEventTypeAdd,
			Peer: peer.ID(),
		})
		// and run protocols
		for _, cap := range caps {
			for _, proto := range srv.Protocols {
				if proto.Name == cap.Name && proto.Version == cap.Version {
					srv.startProtocol(peer, &proto)
				}
			}
		}
	}
	return
}

// Stop terminates the server and all active peer connections.
// It blocks until all active connections have been closed.
func (srv *lachesisServer) Stop() {
	srv.lock.Lock()
	if !srv.running {
		srv.lock.Unlock()
		return
	}

	for _, peer := range srv.peers {
		// broadcast peer drop
		srv.peerFeed.Send(&p2p.PeerEvent{
			Type: p2p.PeerEventTypeDrop,
			Peer: peer.ID(),
		})
	}

	srv.Config.LachesisAdapter.Stop()

	srv.wg.Wait()
	srv.running = false
}

// NodeInfo gathers and returns a collection of metadata known about the host.
func (srv *lachesisServer) NodeInfo() *p2p.NodeInfo {
	// Gather and assemble the generic fake node infos
	node := enode.NewV4(&srv.PrivateKey.PublicKey, net.ParseIP("0.0.0.0"), 0, 0)
	info := &p2p.NodeInfo{
		Name:       srv.Name,
		Enode:      node.String(),
		ID:         node.ID().String(),
		IP:         node.IP().String(),
		ListenAddr: srv.ListenAddr,
		Protocols:  make(map[string]interface{}),
	}
	info.Ports.Discovery = node.UDP()
	info.Ports.Listener = node.TCP()
	if enc, err := rlp.EncodeToBytes(node.Record()); err == nil {
		info.ENR = "0x" + hex.EncodeToString(enc)
	}
	// Gather all the running protocol infos (only once per protocol type)
	for _, proto := range srv.Protocols {
		if _, ok := info.Protocols[proto.Name]; !ok {
			nodeInfo := interface{}("unknown")
			if query := proto.NodeInfo; query != nil {
				nodeInfo = proto.NodeInfo()
			}
			info.Protocols[proto.Name] = nodeInfo
		}
	}

	return info
}

// SubscribePeers subscribes the given channel to peer events.
func (srv *lachesisServer) SubscribeEvents(ch chan *p2p.PeerEvent) event.Subscription {
	return srv.peerFeed.Subscribe(ch)
}

// AddPeer connects to the given node and maintains the connection until the
// server is shut down. If the connection fails for any reason, the server will
// attempt to reconnect the peer.
// Should be empty.
func (srv *lachesisServer) AddPeer(node *enode.Node) {}

// RemovePeer disconnects from the given node.
// Should be empty.
func (srv *lachesisServer) RemovePeer(node *enode.Node) {}

// AddTrustedPeer adds the given node to a reserved whitelist which allows the
// node to always connect, even if the slot are full.
// Should be empty.
func (srv *lachesisServer) AddTrustedPeer(node *enode.Node) {}

// RemoveTrustedPeer removes the given node from the trusted peer set.
// Should be empty.
func (srv *lachesisServer) RemoveTrustedPeer(node *enode.Node) {}

// PeerCount returns the number of connected peers.
func (srv *lachesisServer) PeerCount() int {
	return len(srv.peers)
}

// PeersInfo returns an array of metadata objects describing connected peers.
func (srv *lachesisServer) PeersInfo() []*p2p.PeerInfo {
	// Gather all the generic and sub-protocol specific infos
	infos := make([]*p2p.PeerInfo, 0, srv.PeerCount())
	for _, peer := range srv.peers {
		if peer != nil {
			infos = append(infos, peer.Info())
		}
	}

	return infos
}

// AddProtocols appends the protocols supported
// by the server. Matching protocols are launched for
// each peer.
func (srv *lachesisServer) AddProtocols(protocols ...p2p.Protocol) {
	srv.Protocols = append(srv.Protocols, protocols...)
}

// GetConfig returns server's config
func (srv *lachesisServer) GetConfig() *p2p.Config {
	return &srv.Config
}

// GetDiscV5 returns nil anyway
func (srv *lachesisServer) GetDiscV5() *discv5.Network {
	return nil
}

/*
 * staff
 */

func (srv *lachesisServer) startProtocol(peer *p2p.Peer, proto *p2p.Protocol) {
	srv.log.Debug(fmt.Sprintf("Starting protocol %s-%d for %s", proto.Name, proto.Version, peer.Name()))
	rw := newMsgEventer(srv.Config.LachesisAdapter, &srv.peerFeed, peer.ID(), proto.Name)
	srv.wg.Add(1)
	go func() {
		defer srv.wg.Done()
		err := proto.Run(peer, rw)
		if err == nil {
			srv.log.Trace(fmt.Sprintf("Protocol %s/%d returned", proto.Name, proto.Version))
		} else if err != io.EOF {
			srv.log.Trace(fmt.Sprintf("Protocol %s/%d failed", proto.Name, proto.Version), "err", err)
		} else {
			srv.log.Trace(fmt.Sprintf("Protocol %s/%d closed", proto.Name, proto.Version))
		}
	}()
}

func truncateName(s string) string {
	if len(s) > 20 {
		return s[:20] + "..."
	}
	return s
}
