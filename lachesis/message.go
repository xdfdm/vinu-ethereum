package lachesis

import (
	"io"

	"github.com/ethereum/go-ethereum/p2p"

	"github.com/Fantom-foundation/go-lachesis/src/proxy/proto"
)

/*
 * p2p.MsgReadWriter implementation
 */

// WriteMsg sends a message.
// Can be called simultaneously from multiple goroutines.
func (srv *lachesisServer) ReadMsg() (msg p2p.Msg, err error) {
	select {
	case <-srv.quit:
		err = io.EOF
	case c := <-srv.lachesis.CommitCh():
		err = blockConvert(&c, &msg)
	}

	if err != nil {
		srv.log.Error("ReadMsg", "Err", err)
	}
	return
}

func blockConvert(c *proto.Commit, m *p2p.Msg) error {
	m.Code = 0x07 // eth.NewBlockMsg
	// TODO: convert c.Block.Body.Transactions to m.Payload (ethereum Block)
	m.Size = 0
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
	if msg.Code != 0x02 { // eth/protocol.go: TxMsg
		srv.log.Trace("WriteMsg", "Code", msg.Code)
		return nil
	}
	srv.log.Trace("WriteMsg", "Code", "TxMsg")
	buf := make([]byte, msg.Size)
	_, err = io.ReadFull(msg.Payload, buf)
	if err != nil {
		srv.log.Error("WriteMsg", "Err", err)
		return
	}
	return srv.lachesis.SubmitTx(buf)
}
