package maker

import (
	"cxchain223/blockchain"
	"cxchain223/statdb"
	"cxchain223/statemachine"
	"cxchain223/txpool"
	"cxchain223/types"
	"cxchain223/utils/hash"
	"cxchain223/utils/xtime"
	"math/big"
	"time"
)

type BlockProducerConfig struct {
	Duration   time.Duration
	Difficulty big.Int
	MaxTx      int64
	Coinbase   types.Address
}

type BlockProducer struct {
	txpool txpool.TxPool
	statdb statdb.StatDB
	config BlockProducerConfig

	chain blockchain.Blockchain
	m     statemachine.IMachine

	header *blockchain.Header
	block  *blockchain.Body

	interupt chan bool
}

func (producer BlockProducer) NewBlock() {
	producer.header = blockchain.NewHeader(producer.chain.CurrentHeader)
	producer.header.Coinbase = producer.config.Coinbase
	producer.block = blockchain.NewBlock()
	producer.statdb.SetStatRoot(producer.header.Root)
	// producer.statdb =
}

func (producer BlockProducer) pack() {
	t := time.After(producer.config.Duration)
	txCount := int64(0)
	for {
		select {
		case <-producer.interupt:
			break
		case <-t:
			break
		// TODO 数量
		default:
			if txCount >= producer.config.MaxTx {
				return
			}
			tx := producer.txpool.Pop()
			if tx == nil {
				return
			}
			receiption := producer.m.Execute1(producer.statdb, *tx)
			producer.block.Transactions = append(producer.block.Transactions, *tx)
			producer.block.Receiptions = append(producer.block.Receiptions, *receiption)
			txCount++
		}
	}
}

func (producer BlockProducer) Interupt() {
	producer.interupt <- true
}

func (producer BlockProducer) Seal() (*blockchain.Header, *blockchain.Body) {
	producer.header.Timestamp = xtime.Now()
	producer.header.Nonce = 0

	for {
		hash := producer.header.Hash()

		if meetsDifficulty(hash, producer.config.Difficulty) {
			break
		}
		producer.header.Nonce++
	}

	return producer.header, producer.block
}

func meetsDifficulty(hash hash.Hash, difficulty big.Int) bool {
	difficultyBytes := difficulty.Bytes()
	for i := range difficultyBytes {
		if hash[i] > difficultyBytes[i] {
			return false
		}
	}
	return true
}
