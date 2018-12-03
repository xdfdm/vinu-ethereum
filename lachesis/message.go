package lachesis

import (
	"io"

	"github.com/ethereum/go-ethereum/eth"
	"github.com/ethereum/go-ethereum/event"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/enode"

	"github.com/Fantom-foundation/go-lachesis/src/proxy/proto"
)

/*
 * p2p.MsgReadWriter implementation
 */

// WriteMsg sends a message.
// Can be called simultaneously from multiple goroutines.
func (srv *lachesisServer) ReadMsg() (msg p2p.Msg, err error) {
	c := <-srv.lachesis.CommitCh()
	err = blockConvert(&c, &msg)
	if err != nil {
		// TODO: log
	}
	return
}

func blockConvert(c *proto.Commit, m *p2p.Msg) error {
	m.Code = eth.NewBlockMsg
	// TODO: convert
	return nil
}

// WriteMsg sends a message. It will block until the message's
// Payload has been consumed by the other end.
// Can be called simultaneously from multiple goroutines.
//
// Note that messages can be sent only once because their
// payload reader is drained.
func (srv *lachesisServer) WriteMsg(msg p2p.Msg) (err error) {
	// see codes at eth/protocol.go
	switch msg.Code {
	case eth.StatusMsg:
		srv.log.Trace("WriteMsg", "Code", "StatusMsg")
		break
	case eth.NewBlockHashesMsg:
		srv.log.Trace("WriteMsg", "Code", "NewBlockHashesMsg")
		break
	case eth.TxMsg:
		srv.log.Trace("WriteMsg", "Code", "TxMsg")
		buf := make([]byte, msg.Size)
		_, err = io.ReadFull(msg.Payload, buf)
		if err != nil {
			return
		}
		err = srv.lachesis.SubmitTx(buf)
		if err != nil {
			return
		}
		break
	case eth.GetBlockHeadersMsg:
		srv.log.Trace("WriteMsg", "Code", "GetBlockHeadersMsg")
		break
	case eth.BlockHeadersMsg:
		srv.log.Trace("WriteMsg", "Code", "BlockHeadersMsg")
		break
	case eth.GetBlockBodiesMsg:
		srv.log.Trace("WriteMsg", "Code", "GetBlockBodiesMsg")
		break
	case eth.BlockBodiesMsg:
		srv.log.Trace("WriteMsg", "Code", "BlockBodiesMsg")
		break
	case eth.NewBlockMsg:
		srv.log.Trace("WriteMsg", "Code", "NewBlockMsg")
		break
	}
	return
}

/*
 * p2p.MsgReadWriter wrapper
 */

// msgEventer wraps a p2p.MsgReadWriter and sends events whenever a message is sent
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
