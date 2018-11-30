package lachesis

import (
	"encoding/hex"
	"errors"
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
	peerCount = 5 // see eth.minDesiredPeerCount
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

// lachesisServer manages other peers over lachesis.
type lachesisServer struct {
	// Config fields may not be modified while the server is running.
	p2p.Config

	lock    sync.Mutex // protects running
	running bool

	peers    []*p2p.Peer
	peerFeed event.Feed

	log log.Logger
}

// Start starts running the server.
// Servers can not be re-used after stopping.
func (srv *lachesisServer) Start() error {
	srv.lock.Lock()
	defer srv.lock.Unlock()
	if srv.running {
		return errors.New("server already running")
	}
	srv.running = true

	srv.log = srv.Config.Logger
	if srv.log == nil {
		srv.log = log.New()
	}
	// make fake peers
	// peers should be sorted alphabetically by node identifier
	// (or sort it when PeersInfo())
	for i := 0; i < peerCount; i++ {
		// TODO: replace p2p.Peer with custom to get it's private fields
		/*p := newPeer(c, srv.Protocols)
		// If message events are enabled, pass the peerFeed
		// to the peer
		if srv.EnableMsgEvents {
			p.events = &srv.peerFeed
		}
		name := truncateName(c.name)
		srv.log.Debug("Adding fake peer", "name", name, "addr", c.fd.RemoteAddr(), "peers", len(srv.peers)+1)
		go srv.runPeer(p)
		srv.peers = append(srv.peers, p)
		*/
	}

	return nil
}

// Stop terminates the server and all active peer connections.
// It blocks until all active connections have been closed.
func (srv *lachesisServer) Stop() {
	srv.lock.Lock()
	if !srv.running {
		srv.lock.Unlock()
		return
	}
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

// runPeer runs in its own goroutine for each peer.
// it waits until the Peer logic returns and removes
// the peer.
func (srv *lachesisServer) runPeer(p *p2p.Peer) {
	// broadcast peer add
	srv.peerFeed.Send(&p2p.PeerEvent{
		Type: p2p.PeerEventTypeAdd,
		Peer: p.ID(),
	})

	// run the protocol
	var err error
	/*
		remoteRequested, err := p.run()
	*/

	// broadcast peer drop
	srv.peerFeed.Send(&p2p.PeerEvent{
		Type:  p2p.PeerEventTypeDrop,
		Peer:  p.ID(),
		Error: err.Error(),
	})
}

func truncateName(s string) string {
	if len(s) > 20 {
		return s[:20] + "..."
	}
	return s
}
