package maker

import (
	"goXChain/blockchain"
	"goXChain/statdb"
	"goXChain/statemachine"
	"goXChain/txpool"
	"goXChain/types"
	"time"
)

type ChainConfig struct {
	Duration   time.Duration
	Coinbase   types.Address
	Difficulty uint
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

func NewBlockMaker(txpool txpool.TxPool, state statdb.StatDB, exec statemachine.StateMachine, config ChainConfig) *BlockMaker {
	return &BlockMaker{
		txpool: txpool,
		state:  state,
		exec:   exec,
		config: config,
	}
}

func (maker BlockMaker) NewBlock(header blockchain.Header) blockchain.Blockchain {
	maker.nextBody = blockchain.NewBody()
	maker.Pack()
	maker.nextHeader = blockchain.NewHeader(header, maker.state)
	maker.nextHeader, maker.nextBody = maker.Finalize()
	maker.nextHeader.Coinbase = maker.config.Coinbase
	maker.chain.CurrentHeader = *maker.nextHeader
	maker.chain.State = maker.state.All()
	maker.chain.Txpool = maker.txpool
	return maker.chain
}

func (maker BlockMaker) Pack() {
	end := time.After(maker.config.Duration)
	for {
		select {
		case <-maker.interupt:
			return
		case <-end:
			return
		default:
			// 如果没有接收到中断信号和定时器触发信号，则执行其他操作（这里假设是 maker.pack()）
			<-end
			maker.pack()
			return
		}
	}
	maker.pack()
}

func (maker BlockMaker) pack() {
	tx := maker.txpool.Pop()
	if tx == nil {
		return
	}
	receiption := maker.exec.Execute1(maker.state, *tx)
	maker.nextBody.Transactions = append(maker.nextBody.Transactions, *tx)
	maker.nextBody.Receiptions = append(maker.nextBody.Receiptions, *receiption)
}

func (maker BlockMaker) Interrupt() {
	maker.interupt <- true
}

func (maker *BlockMaker) Finalize() (*blockchain.Header, *blockchain.Body) {
	maker.nextHeader.Timestamp = uint64(time.Now().Unix())
	maker.nextHeader.Nonce = 0
	// 匹配nonce
	for {
		maker.nextHeader.Nonce++
		hash := maker.nextHeader.Hash()
		if hash.Cmp(maker.chain.CurrentHeader.Hash()) < 0{
			break
		}
	}
	return maker.nextHeader, maker.nextBody
}
