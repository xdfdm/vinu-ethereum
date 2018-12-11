package eth

import (
	"bytes"
	"io"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/rlp"

	"github.com/Fantom-foundation/go-lachesis/src/proxy"
	"github.com/Fantom-foundation/go-lachesis/src/proxy/proto"
)

func NewLachesisAdapter(addr string) p2p.LachesisAdapter {
	return &lachesisAdapter{
		addr: addr,
	}
}

type lachesisAdapter struct {
	addr string

	log log.Logger

	lachesis proxy.LachesisProxy
	quit     chan struct{}
}

/*
 * p2p.LachesisAdapter implementation
 */

func (srv *lachesisAdapter) Start(log log.Logger) (err error) {
	srv.log = log
	srv.log.Debug("lachesisAdapter.Start()")

	srv.quit = make(chan struct{})

	srv.lachesis, err = proxy.NewGrpcLachesisProxy(srv.addr, nil)
	if err != nil {
		return
	}

	return
}

func (srv *lachesisAdapter) Stop() {
	close(srv.quit)
}

func (srv *lachesisAdapter) Address() string {
	return srv.addr
}

// ReadMsg returns a message.
// Can be called simultaneously from multiple goroutines.
func (srv *lachesisAdapter) ReadMsg() (msg p2p.Msg, err error) {
	srv.log.Debug("lachesisAdapter.ReadMsg")
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
	srv.log.Debug("lachesisAdapter.WriteMsg", "msg", msg)
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
	// parse txs
	var txs types.Transactions
	for _, raw := range c.Block.Body.Transactions {
		tx := &types.Transaction{}
		err := tx.DecodeRLP(rlp.NewStream(bytes.NewReader(raw), uint64(len(raw))))
		if err != nil {
			return err
		}
		txs = append(txs, tx)
	}

	// make block
	// TODO: fill all
	header := &types.Header{}
	var uncles []*types.Header
	var receipts []*types.Receipt
	block := types.NewBlock(header, txs, uncles, receipts)

	// block to Msg
	size, r, err := rlp.EncodeToReader(block)
	if err != nil {
		return err
	}
	m.Code = NewBlockMsg
	m.Size = uint32(size)
	m.Payload = r

	return nil
}
