package p2p

import (
	"github.com/ethereum/go-ethereum/event"
	"github.com/ethereum/go-ethereum/p2p/discv5"
	"github.com/ethereum/go-ethereum/p2p/enode"
)

/*
 * Interface
 */

type ServerInterface interface {
	// Start starts running the server.
	// Servers can not be re-used after stopping.
	Start() error
	// Stop terminates the server and all active peer connections.
	// It blocks until all active connections have been closed.
	Stop()

	// NodeInfo gathers and returns a collection of metadata known about the host.
	NodeInfo() *NodeInfo

	// SubscribePeers subscribes the given channel to peer events.
	SubscribeEvents(ch chan *PeerEvent) event.Subscription

	// AddPeer connects to the given node and maintains the connection until the
	// server is shut down. If the connection fails for any reason, the server will
	// attempt to reconnect the peer.
	AddPeer(node *enode.Node)
	// RemovePeer disconnects from the given node.
	RemovePeer(node *enode.Node)

	// AddTrustedPeer adds the given node to a reserved whitelist which allows the
	// node to always connect, even if the slot are full.
	AddTrustedPeer(node *enode.Node)
	// RemoveTrustedPeer removes the given node from the trusted peer set.
	RemoveTrustedPeer(node *enode.Node)

	// PeerCount returns the number of connected peers.
	PeerCount() int
	// PeersInfo returns an array of metadata objects describing connected peers.
	PeersInfo() []*PeerInfo

	// AddProtocols appends the protocols supported
	// by the server. Matching protocols are launched for
	// each peer.
	AddProtocols(protocols ...Protocol)

	// GetConfig returns server's config
	GetConfig() *Config

	// GetDiscV5 returns discV5 Network if set
	GetDiscV5() *discv5.Network
}

/*
 * Wrapper
 */

func NewServer(cfg Config) *Server {
	return WrapServer(&p2pServer{Config: cfg})
}

// Server is struct wrapper for ServerInterface.
// Used to don't change ethereum code
type Server struct {
	I ServerInterface
	Config
	DiscV5 *discv5.Network
}

// WrapServer wraps ServerInterface into struct
func WrapServer(s ServerInterface) *Server {
	return &Server{
		I: s,
	}
}

// Start starts running the server.
// Servers can not be re-used after stopping.
func (s *Server) Start() error {
	err := s.I.Start()
	// refresh fields after start
	s.Config = *s.I.GetConfig()
	s.DiscV5 = s.I.GetDiscV5()

	return err
}

// Stop terminates the server and all active peer connections.
// It blocks until all active connections have been closed.
func (s *Server) Stop() {
	s.I.Stop()
}

// NodeInfo gathers and returns a collection of metadata known about the host.
func (s *Server) NodeInfo() *NodeInfo {
	return s.I.NodeInfo()
}

// SubscribePeers subscribes the given channel to peer events.
func (s *Server) SubscribeEvents(ch chan *PeerEvent) event.Subscription {
	return s.I.SubscribeEvents(ch)
}

// AddPeer connects to the given node and maintains the connection until the
// server is shut down. If the connection fails for any reason, the server will
// attempt to reconnect the peer.
func (s *Server) AddPeer(node *enode.Node) {
	s.I.AddPeer(node)
}

// RemovePeer disconnects from the given node.
func (s *Server) RemovePeer(node *enode.Node) {
	s.I.RemovePeer(node)
}

// AddTrustedPeer adds the given node to a reserved whitelist which allows the
// node to always connect, even if the slot are full.
func (s *Server) AddTrustedPeer(node *enode.Node) {
	s.I.AddTrustedPeer(node)
}

// RemoveTrustedPeer removes the given node from the trusted peer set.
func (s *Server) RemoveTrustedPeer(node *enode.Node) {
	s.I.RemoveTrustedPeer(node)
}

// PeerCount returns the number of connected peers.
func (s *Server) PeerCount() int {
	return s.I.PeerCount()
}

// PeersInfo returns an array of metadata objects describing connected peers.
func (s *Server) PeersInfo() []*PeerInfo {
	return s.I.PeersInfo()
}

// AddProtocols appends the protocols supported
// by the server. Matching protocols are launched for
// each peer.
func (s *Server) AddProtocols(protocols ...Protocol) {
	s.I.AddProtocols(protocols...)
}

/*
 * ServerImplementation
 */

// AddProtocols appends the protocols supported
// by the server. Matching protocols are launched for
// each peer.
func (srv *p2pServer) AddProtocols(protocols ...Protocol) {
	srv.Protocols = append(srv.Protocols, protocols...)
}

// GetConfig returns server config
func (srv *p2pServer) GetConfig() *Config {
	return &srv.Config
}

// GetDiscV5 returns discV5 Network
func (srv *p2pServer) GetDiscV5() *discv5.Network {
	return srv.DiscV5
}
