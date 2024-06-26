package maker

import (
	"cxchain223/blockchain"
	"cxchain223/statdb"
	"cxchain223/statemachine"
	"cxchain223/txpool"
	"cxchain223/types"
	"cxchain223/utils/xtime"
	"encoding/binary"
	"time"
)

type ChainConfig struct {
	Duration   time.Duration
	Coinbase   types.Address
	Difficulty uint64
}

type BlockMaker struct {
	txpool txpool.TxPool
	state  statdb.StatDB
	exec   statemachine.IMachine

	config ChainConfig
	chain  blockchain.Blockchain

	nextHeader *blockchain.Header
	nextBody   *blockchain.Body

	interupt chan bool
}

// NewBlockMaker init BlockMaker
func NewBlockMaker(txpool txpool.TxPool, state statdb.StatDB, exec statemachine.StateMachine) *BlockMaker {
	return &BlockMaker{
		txpool: txpool,
		state:  state,
		exec:   exec,
	}
}

// NewBlock  BlockMaker calls NewBlock to make a  complete block
func (maker BlockMaker) NewBlock() {
	maker.nextBody = blockchain.NewBlock()
	maker.nextHeader = blockchain.NewHeader(maker.chain.CurrentHeader)
	maker.nextHeader.Coinbase = maker.config.Coinbase
}

func (maker BlockMaker) Pack() {
	end := time.After(maker.config.Duration)
	for {
		select {
		case <-maker.interupt:
			break
		case <-end:
			break
		default:
			maker.pack()
		}
	}
}

func (maker BlockMaker) pack() {
	tx := maker.txpool.Pop()
	receiption := maker.exec.Execute1(maker.state, *tx)
	maker.nextBody.Transactions = append(maker.nextBody.Transactions, *tx)
	maker.nextBody.Receiptions = append(maker.nextBody.Receiptions, *receiption)
}

func (maker BlockMaker) Interupt() {
	maker.interupt <- true
}

func (maker BlockMaker) Finalize() (*blockchain.Header, *blockchain.Body) {
	maker.nextHeader.Timestamp = xtime.Now()
	maker.nextHeader.Nonce = 0
	// TODO
	for n := uint64(0); ; n++ {
		maker.nextHeader.Nonce = n
		hash := maker.nextHeader.Hash()
		_hash := binary.BigEndian.Uint64(hash[:8]) //hash.Hash transfer to uint64
		if _hash < maker.config.Difficulty {
			break
		}
	}

	return maker.nextHeader, maker.nextBody
}
