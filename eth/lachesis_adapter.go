package eth

import (
	"fmt"
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
	addr     string
	log      log.Logger
	answers  chan *p2p.Msg
	lachesis proxy.LachesisProxy
	quit     chan struct{}
}

/*
 * p2p.LachesisAdapter implementation
 */

func (srv *lachesisAdapter) Start(log log.Logger) (err error) {
	srv.log = log
	srv.log.Debug("lachesisAdapter.Start()")

	srv.answers = make(chan *p2p.Msg, 5) // max peer count
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
	for {
		select {
		case <-srv.quit:
			srv.log.Debug("lachesisAdapter.ReadMsg quit")
			err = io.EOF
			return
		case m := <-srv.answers:
			srv.log.Debug("lachesisAdapter.ReadMsg answer", "msg", msg)
			msg = *m
			err = nil
			return
		case c := <-srv.lachesis.CommitCh():
			err = srv.blockConvert(&c, &msg)
			if err == nil {
				srv.log.Debug("lachesisAdapter.ReadMsg commit", "block", c.Block.Body)
				return
			} else {
				srv.log.Warn("lachesisAdapter.ReadMsg commit", "block", c.Block.Body, "err", err)
				continue
			}
		}
	}
}

// WriteMsg sends a message. It will block until the message's
// Payload has been consumed by the other end.
// Can be called simultaneously from multiple goroutines.
//
// Note that messages can be sent only once because their
// payload reader is drained.
func (srv *lachesisAdapter) WriteMsg(msg p2p.Msg) (err error) {
	switch msg.Code {

	case StatusMsg:
		srv.log.Debug("lachesisAdapter.WriteMsg handshake", "msg", msg)
		srv.sendAnswer(&msg)
		return nil

	case TxMsg:
		srv.log.Debug("lachesisAdapter.WriteMsg tx", "msg", msg)
		var txs []*types.Transaction
		if err := msg.Decode(&txs); err != nil {
			return errResp(ErrDecode, "msg %v: %v", msg, err)
		}
		for _, tx := range txs {
			buf, err := rlp.EncodeToBytes(tx)
			if err != nil {
				return err
			}
			err = srv.lachesis.SubmitTx(buf)
			if err != nil {
				return err
			}
		}
		return nil

	default:
		srv.log.Debug("lachesisAdapter.WriteMsg skip", "code", msg.Code)
		return nil

	}
}

/*
 * staff
 */

// sendAnswer sends data for ReadMsg()
func (srv *lachesisAdapter) sendAnswer(msg *p2p.Msg) {
	srv.answers <- msg
}

func (srv *lachesisAdapter) blockConvert(c *proto.Commit, m *p2p.Msg) error {
	// parse txs
	var txs []*types.Transaction
	for _, raw := range c.Block.Body.Transactions {
		tx := &types.Transaction{}
		err := rlp.DecodeBytes(raw, tx)
		if err != nil {
			srv.log.Warn("invalid tx in Commit", "err", err)
		} else {
			txs = append(txs, tx)
		}
	}

	if len(txs) < 1 {
		return fmt.Errorf("no valid txs")
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
