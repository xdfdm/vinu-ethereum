package lachesis

import (
	"io"

	"github.com/ethereum/go-ethereum/event"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/enode"
)

// msgEventer is p2p.MsgReadWriter and  wraps a p2p.MsgReadWriter and sends events whenever a message is sent
// or received
type msgEventer struct {
	p2p.MsgReadWriter

	feed     *event.Feed
	peerID   enode.ID
	Protocol string
}

// newMsgEventer returns a msgEventer which sends message events to the given
// feed
func newMsgEventer(rw p2p.MsgReadWriter, feed *event.Feed, peerID enode.ID, proto string) *msgEventer {
	return &msgEventer{
		MsgReadWriter: rw,
		feed:          feed,
		peerID:        peerID,
		Protocol:      proto,
	}
}

// ReadMsg reads a message from the underlying p2p.MsgReadWriter and emits a
// "message received" event
func (ev *msgEventer) ReadMsg() (p2p.Msg, error) {
	msg, err := ev.MsgReadWriter.ReadMsg()
	if err != nil {
		return msg, err
	}
	ev.feed.Send(&p2p.PeerEvent{
		Type:     p2p.PeerEventTypeMsgRecv,
		Peer:     ev.peerID,
		Protocol: ev.Protocol,
		MsgCode:  &msg.Code,
		MsgSize:  &msg.Size,
	})
	return msg, nil
}

// WriteMsg writes a message to the underlying p2p.MsgReadWriter and emits a
// "message sent" event
func (ev *msgEventer) WriteMsg(msg p2p.Msg) error {
	err := ev.MsgReadWriter.WriteMsg(msg)
	if err != nil {
		return err
	}
	ev.feed.Send(&p2p.PeerEvent{
		Type:     p2p.PeerEventTypeMsgSend,
		Peer:     ev.peerID,
		Protocol: ev.Protocol,
		MsgCode:  &msg.Code,
		MsgSize:  &msg.Size,
	})
	return nil
}

// Close closes the underlying p2p.MsgReadWriter if it implements the io.Closer
// interface
func (ev *msgEventer) Close() error {
	if v, ok := ev.MsgReadWriter.(io.Closer); ok {
		return v.Close()
	}
	return nil
}
