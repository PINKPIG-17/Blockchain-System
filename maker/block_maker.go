package maker

import (
	"cxchain223/blockchain"
	"cxchain223/statdb"
	"cxchain223/statemachine"
	"cxchain223/trie"
	"cxchain223/txpool"
	"cxchain223/utils/rlp"
	"cxchain223/utils/xtime"
	"encoding/binary"
	"fmt"
	"time"
)

type BlockMaker struct {
	txpool txpool.TxPool
	state  statdb.StatDB
	exec   statemachine.IMachine

	config blockchain.ChainConfig
	chain  *blockchain.Blockchain

	nextHeader *blockchain.Header
	nextBody   *blockchain.Body

	interupt chan bool
	txDB     trie.ITrie
}

func NewBlockMaker(txpool *txpool.DefaultPool, state *statdb.IStatDB, exec *statemachine.StateMachine, chain *blockchain.Blockchain, txdb trie.ITrie) *BlockMaker {
	return &BlockMaker{
		txpool:     txpool,
		state:      state,
		exec:       exec,
		chain:      chain,
		config:     blockchain.DefaultConfig,
		txDB:       txdb,
		nextHeader: &blockchain.EmptyHeader,
	}
}

func (maker *BlockMaker) NewBlock() {
	maker.nextBody = blockchain.NewBlock()
	maker.nextHeader = blockchain.NewHeader(maker.chain.CurrentHeader)
	maker.nextHeader.Coinbase = maker.config.Coinbase
}

func (maker *BlockMaker) Pack() {
	end := time.After(maker.config.Duration)
	for {
		select {
		case <-maker.interupt:
			break
		case <-end:
			break
		default:
			if maker.txpool.Has() {
				maker.pack()
			}
		}
	}
}

func (maker *BlockMaker) pack() {
	tx := maker.txpool.Pop()
	receiption := maker.exec.Execute1(maker.state, *tx)
	maker.nextBody.Transactions = append(maker.nextBody.Transactions, *tx)
	maker.nextBody.Receiptions = append(maker.nextBody.Receiptions, *receiption)
}

func (maker *BlockMaker) Interupt() {
	maker.interupt <- true
}

func (maker *BlockMaker) Finalize() (*blockchain.Header, *blockchain.Body) {
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
func (maker *BlockMaker) Seal() {
	maker.NewBlock()
	for {
		if maker.txpool.Has() {
			maker.pack()
		} else {
			if len(maker.nextBody.Transactions) > 0 {
				blk := blockchain.Block{
					*maker.nextHeader,
					*maker.nextBody,
				}
				str := blk.ToString()
				err := maker.txDB.Store(blk.Root[:], []byte(str))
				if err != nil {
					panic(err)
				}
				fmt.Println("Generate Block:", maker.nextHeader.Height, "\nroot:", maker.nextHeader.Root)
				maker.chain.CurrentHeader = &blk.Header
				maker.NewBlock()
				fmt.Println("The Next Block:", maker.nextHeader.Root)
			}
			return
		}
	}

}
func (maker *BlockMaker) SealWithTime(t time.Duration) {
	go maker.Seal()
	ticker := time.NewTicker(t)
	for maker.txpool.Has() {
		select {
		case <-ticker.C:
			maker.Interupt()
			if len(maker.nextBody.Transactions) > 0 {
				blk := blockchain.Block{
					*maker.nextHeader,
					*maker.nextBody,
				}
				bts, err := rlp.EncodeToBytes(blk)
				if err != nil {
					panic(err)
				}
				err = maker.txDB.Store(blk.Root[:], bts)
				if err != nil {
					panic(err)
				}
				maker.chain.CurrentHeader = &blk.Header
				maker.NewBlock()
				fmt.Println("Generate Block:", maker.nextHeader.Height)
			}
			go maker.Seal()
		}
	}
}
