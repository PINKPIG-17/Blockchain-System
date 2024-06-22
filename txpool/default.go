package txpool

import (
	"cxchain223/statdb"
	"cxchain223/types"
	"cxchain223/utils/hash"
	"sort"
)

type SortedTxs interface {
	GasPrice() uint64
	Push(tx *types.Transaction)
	Replace(tx *types.Transaction)
	Pop() *types.Transaction
	Nonce() uint64
	Length() int
}

type DefaultSortedTxs []*types.Transaction

func (sorted DefaultSortedTxs) GasPrice() uint64 {
	first := sorted[0]
	return first.GasPrice
}

func (sorted DefaultSortedTxs) Nonce() uint64 {
	last := sorted[len(sorted)-1]
	return last.Nonce
}

func (sorted DefaultSortedTxs) Push(tx *types.Transaction) {
	sorted = append(sorted, tx)
}

func (sorted DefaultSortedTxs) Pop() *types.Transaction {
	pop := sorted[len(sorted)-1]
	sorted = sorted[:len(sorted)-1]
	return pop
}

func (sorted DefaultSortedTxs) Replace(tx *types.Transaction) {
	from := tx.From()
	for i := 0; i < len(sorted)-1; i++ {
		if sorted[i].From() == from {
			if sorted[i].Nonce == tx.Nonce {
				sorted[i] = tx
			}
		}
	}
}

func (sorted DefaultSortedTxs) Length() int {
	return len(sorted)
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

func (pool DefaultPool) SetStatRoot(root hash.Hash) {
	pool.Stat.SetStatRoot(root)
}

func (pool DefaultPool) NewTx(tx *types.Transaction) {
	account := pool.Stat.Load(tx.From())
	if account.Nonce >= tx.Nonce {
		return
	}

	nonce := account.Nonce
	blks := pool.pendings[tx.From()]
	if len(blks) > 0 {
		last := blks[len(blks)-1]
		nonce = last.Nonce()
	}
	if tx.Nonce > nonce+1 {
		pool.addQueueTx(tx)
	} else if tx.Nonce == nonce+1 {
		// push
		pool.pushPendingTx(blks, tx)
	} else {
		// 替换
		pool.replacePendingTx(blks, tx)
	}
}

func (pool DefaultPool) replacePendingTx(blks []SortedTxs, tx *types.Transaction) {
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

func (pool DefaultPool) pushPendingTx(blks []SortedTxs, tx *types.Transaction) {
	if len(blks) == 0 {
		blk := make(DefaultSortedTxs, 1)
		blk = append(blk, tx)
		blks = append(blks, blk)
		pool.pendings[tx.From()] = blks
		pool.txs = append(pool.txs, blk)
		sort.Sort(pool.txs)
	} else {
		last := blks[len(blks)-1]
		if last.GasPrice() <= tx.GasPrice {
			last.Push(tx)
		} else {
			blk := make(DefaultSortedTxs, 1)
			blk = append(blk, tx)
			blks = append(blks, blk)
			pool.pendings[tx.From()] = blks
			pool.txs = append(pool.txs, blk)
			sort.Sort(pool.txs)
		}
	}
}

func (pool DefaultPool) addQueueTx(tx *types.Transaction) {
	list := pool.queue[tx.From()]
	list = append(list, tx)
	sort.Slice(list, func(i, j int) bool {
		return list[i].Nonce < list[j].Nonce
	})
	pool.queue[tx.From()] = list
}
func (pool *DefaultPool) Pop() *types.Transaction {
	if (*pool).txs[0].Length() == 0 {
		(*pool).txs = (*pool).txs[0:]
	}
	return (*pool).txs[0].Pop()
}

func (pool DefaultPool) NotifyTxEvent(txs []*types.Transaction) {

}
