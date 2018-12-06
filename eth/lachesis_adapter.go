package eth

import (
	"io"

	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/p2p"

	"github.com/Fantom-foundation/go-lachesis/src/proxy"
	"github.com/Fantom-foundation/go-lachesis/src/proxy/proto"
)

func NewLachesisAdapter(addr string, log log.Logger) p2p.LachesisAdapter {
	return &lachesisAdapter{
		Addr: addr,
		log:  log,
	}
}

type lachesisAdapter struct {
	Addr string

	log log.Logger

	lachesis proxy.LachesisProxy
	quit     chan struct{}
}

/*
 * p2p.LachesisAdapter implementation
 */

func (srv *lachesisAdapter) Start() (err error) {
	srv.quit = make(chan struct{})

	srv.lachesis, err = proxy.NewGrpcLachesisProxy(srv.Addr, nil)
	if err != nil {
		return
	}

	return
}

func (srv *lachesisAdapter) Stop() {
	close(srv.quit)
}

func (srv *lachesisAdapter) Address() string {
	return srv.Addr
}

// ReadMsg returns a message.
// Can be called simultaneously from multiple goroutines.
func (srv *lachesisAdapter) ReadMsg() (msg p2p.Msg, err error) {
	// TODO: make it clever
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

// WriteMsg sends a message. It will block until the message's
// Payload has been consumed by the other end.
// Can be called simultaneously from multiple goroutines.
//
// Note that messages can be sent only once because their
// payload reader is drained.
func (srv *lachesisAdapter) WriteMsg(msg p2p.Msg) (err error) {
	// TODO: make it clever
	if msg.Code != TxMsg {
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

/*
 * staff
 */

func blockConvert(c *proto.Commit, m *p2p.Msg) error {
	m.Code = NewBlockMsg
	// TODO: convert c.Block.Body.Transactions to m.Payload (eth Block)
	m.Size = 0
	return nil
}
