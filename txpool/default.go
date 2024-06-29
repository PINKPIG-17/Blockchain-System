package txpool

import (
	"crypto/ecdsa"
	"goXChain/crypto/sha3"
	"goXChain/statdb"
	"goXChain/types"
	"goXChain/utils/hash"
	"goXChain/utils/rlp"
	"fmt"
	"sort"
)

type SortedTxs interface {
	GasPrice() uint64
	Push(tx *types.Transaction)
	Replace(tx *types.Transaction)
	Pop() *types.Transaction
	Nonce() uint64
}

type DefaultSortedTxs []*types.Transaction

func (sorted DefaultSortedTxs) GasPrice() uint64 {
	first := sorted[len(sorted)-1]
	return first.GasPrice
}

func (sorted *DefaultSortedTxs) Push(tx *types.Transaction) {
	*sorted = append(*sorted, tx)
	sort.Slice(*sorted, func(i, j int) bool {
		return (*sorted)[i].Nonce < (*sorted)[j].Nonce
	})
}

func (sorted *DefaultSortedTxs) Replace(tx *types.Transaction) {
	replaced := false

	// 遍历 sorted 中的交易，找到 Nonce 相同的交易
	for i, t := range *sorted {
		if t.Nonce == tx.Nonce {
			// 如果找到相同的 Nonce，比较 GasPrice 并替换为新的交易
			if tx.GasPrice > t.GasPrice {
				(*sorted)[i] = tx
			}
			replaced = true
			break
		}
	}

	// 如果未找到相同的 Nonce，插入交易并排序
	if !replaced {
		sorted.Push(tx)
	}
}

func (sorted *DefaultSortedTxs) Pop() *types.Transaction {
	//f := (*sorted)[len(*sorted)-1]
	//*sorted = (*sorted)[:len(*sorted)-1]
	//return f
	if len(*sorted) == 0 {
		return nil
	}
	if (*sorted)[0] == nil {
		*sorted = (*sorted)[1:]
	}
	first := (*sorted)[0]
	*sorted = (*sorted)[1:]
	return first
}

func (sorted DefaultSortedTxs) Nonce() uint64 {
	if len(sorted) == 0 {
		return 0
	}
	return sorted[len(sorted)-1].Nonce
}

type pendingTxs []SortedTxs

func (p pendingTxs) Len() int {
	return len(p)
}

func (p pendingTxs) Less(i, j int) bool {
	return p[i].GasPrice() < p[j].GasPrice()
}

func (p pendingTxs) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

type DefaultPool struct {
	Stat statdb.StatDB

	all      map[hash.Hash]bool
	txs      pendingTxs
	pendings map[types.Address][]SortedTxs
	queue    map[types.Address][]*types.Transaction
}

func NewDefaultPool(stat statdb.StatDB) *DefaultPool {
	return &DefaultPool{
		Stat:     stat,
		all:      make(map[hash.Hash]bool),
		txs:      make(pendingTxs, 0),
		pendings: make(map[types.Address][]SortedTxs),
		queue:    make(map[types.Address][]*types.Transaction),
	}
}

func (pool *DefaultPool) replacePendingTx(blks []SortedTxs, tx *types.Transaction) {
	for _, blk := range blks {
		if blk.Nonce() >= tx.Nonce {
			// replace
			if blk.GasPrice() <= tx.GasPrice {
				blk.Replace(tx)
			}
			break
		}
	}
}

func (pool *DefaultPool) pushPendingTx(blks []SortedTxs, tx *types.Transaction) {
	if len(blks) == 0 {
		blk := make(DefaultSortedTxs, 1)
		blk = append(blk, tx)
		blks = append(blks, &blk)
		pool.pendings[tx.Txdata.From] = blks
		pool.txs = append(pool.txs, &blk)
		sort.Sort(pool.txs)
	} else {
		last := blks[len(blks)-1]
		if last.GasPrice() <= tx.GasPrice {
			last.Push(tx)
		} else {
			blk := make(DefaultSortedTxs, 1)
			blk = append(blk, tx)
			blks = append(blks, &blk)
			pool.pendings[tx.Txdata.From] = blks
			pool.txs = append(pool.txs, &blk)
			sort.Sort(pool.txs)
		}
	}
}

func (pool *DefaultPool) addQueueTx(tx *types.Transaction) {
	address := tx.Txdata.From
	list := pool.queue[address]
	list = append(list, tx)
	// 对列表中的交易按照 GasPrice 进行排序
	sort.Slice(list, func(i, j int) bool {
		return list[i].GasPrice > list[j].GasPrice
	})
	pool.queue[address] = list
}

func (pool *DefaultPool) SetStatRoot(root hash.Hash) {
	pool.Stat.SetStatRoot(root)
}

func (pool *DefaultPool) NewTx(tx *types.Transaction, pubk ecdsa.PublicKey) {
	bytes, _ := rlp.EncodeToBytes(tx.Txdata)
	if !types.VerifySignature(&pubk, bytes, &tx.Signature) {
		fmt.Println("this signature not right")
		return
	}
	pubkbyte, _ := rlp.EncodeToBytes(pubk)
	if types.PubKeyToAddress(pubkbyte) != tx.Txdata.From {
		fmt.Println("you are not owner")
		return
	}

	account := pool.Stat.Load(tx.Txdata.From)
	if account.Nonce >= tx.Nonce {
		return
	}

	txHash := sha3.Keccak256(bytes)

	// 检查交易是否已经存在于 all 中
	if pool.all[txHash] {
		fmt.Println("transaction already exists in the pool")
		return
	}

	nonce := account.Nonce + 1
	blks := pool.pendings[tx.Txdata.From]
	if len(blks) > 0 {
		last := blks[len(blks)-1]
		nonce = last.Nonce()
	}
	if tx.Nonce > nonce {
		pool.addQueueTx(tx)
	} else if tx.Nonce == nonce {
		// push
		pool.pushPendingTx(blks, tx)
	} else {
		// 替换
		pool.replacePendingTx(blks, tx)
	}
	pool.all[txHash] = true
	// 检查 queue 中是否有可以处理的交易
	pool.processQueue(tx.Txdata.From)
}

func (pool *DefaultPool) processQueue(address types.Address) {
	if txs, ok := pool.queue[address]; ok {
		account := pool.Stat.Load(address)
		nonce := account.Nonce + 1

		// 获取当前的 pendings
		blks := pool.pendings[address]
		if len(blks) > 0 {
			last := blks[len(blks)-1]
			nonce = last.Nonce()
		}

		// 逐个检查 queue 中的交易
		newQueue := []*types.Transaction{}
		for _, tx := range txs {
			if tx.Nonce == nonce {
				pool.pushPendingTx(blks, tx)
				bytes, _ := rlp.EncodeToBytes(tx)
				pool.all[sha3.Keccak256(bytes)] = true
				nonce++
			} else if tx.Nonce > nonce {
				newQueue = append(newQueue, tx)
			}
		}

		// 更新 queue
		if len(newQueue) == 0 {
			delete(pool.queue, address)
		} else {
			pool.queue[address] = newQueue
		}
	}
}

func (pool *DefaultPool) Pop() *types.Transaction {
	if len(pool.txs) == 0 {
		return nil
	}
	// 从排序的交易池中弹出第一个交易
	firstBlk := pool.txs[0]
	tx := firstBlk.Pop()
	pool.txs = pool.txs[1:]
	// 计算交易的哈希值
	bytes, _ := rlp.EncodeToBytes(tx.Txdata)
	txHash := sha3.Keccak256(bytes)

	// 从 all 中移除交易的哈希值
	delete(pool.all, txHash)

	return tx
}
func (pool *DefaultPool) NotifyTxEvent(txs []*types.Transaction, pubk []ecdsa.PublicKey) {
	for index, tx := range txs {
		pool.NewTx(tx, pubk[index])
	}
}
