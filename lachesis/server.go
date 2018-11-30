package lachesis

import (
	"encoding/hex"
	"net"

	"github.com/ethereum/go-ethereum/event"
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

	peers []*p2p.Peer
}

// Start starts running the server.
// Servers can not be re-used after stopping.
func (srv *lachesisServer) Start() error {
	// TODO: implement
	return nil
}

// Stop terminates the server and all active peer connections.
// It blocks until all active connections have been closed.
func (srv *lachesisServer) Stop() {
	// TODO: implement
}

// NodeInfo gathers and returns a collection of metadata known about the host.
func (srv *lachesisServer) NodeInfo() *p2p.NodeInfo {
	// Gather and assemble the generic node infos
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
	// TODO: implement
	return nil
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
	// TODO: implement
	return make([]*p2p.PeerInfo, 0)
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
