package main

import (
	"cxchain223/kvstore"
	"cxchain223/trie"
)

func main() {
	db := kvstore.NewLevelDB("./testdb")
	state := trie.NewState(db, trie.EmptyHash)
	state.Store([]byte("apple"), []byte("apple"))
	state.Store([]byte("apply"), []byte("apply"))
	state.Store([]byte("application"), []byte("application"))
	state.Store([]byte("banana"), []byte("banana"))
	state.Store([]byte("band"), []byte("band"))
	// value, err := state.Load([]byte("apple"))
	// values, err := state.Load([]byte("apply"))
	// fmt.Println(string(values), err)
	// fmt.Println(string(value), err)

	// statedbRoot := statdb.statdbroot{state: state}
	// defaultpool := txpool.DefaultPool{Stat: statedbRoot, all: make(map[hash.Hash]bool), txs: make([]txpool.SortedTxs, 0), pendings: make(map[types.Address]txpool.DefaultSortedTxs), queue: make(map[types.Address]*types.Transaction)}

	// txpool := txpool.DefaultPool()
	// statdb := statdb.StatDB() // 自行实现一个模拟的 StatDB 接口

	// producer := maker.BlockProducer{
	// 	TxPool: txpool,
	// 	StatDB: statdb,
	// 	Config: maker.BlockProducerConfig{
	// 		Duration:   5 * time.Second,  // 每5秒确认一个区块
	// 		Difficulty: *big.NewInt(100), // 替换为实际的难度值
	// 		MaxTx:      100,
	// 		Coinbase:   types.Address{}, // 替换为实际的地址
	// 	},
	// 	Chain:    blockchain.Blockchain{}, // 如果有现成的区块链对象，替换为实际的区块链对象
	// 	M:        maker.StateMachine{},    // 替换为实际的状态机对象
	// 	Header:   &blockchain.Header{},    // 如果有现成的区块头对象，替换为实际的区块头对象
	// 	Block:    &blockchain.Body{},      // 如果有现成的区块体对象，替换为实际的区块体对象
	// 	Interupt: make(chan bool),         // 替换为实际的通道对象
	// }

	// // 调用 Seal 方法
	// header, block := producer.Seal()

	// // 打印生成的区块头和区块体
	// fmt.Println("Generated Header:", header)
	// fmt.Println("Generated Block:", block)
}
