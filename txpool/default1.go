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
}
type arr_tx []DefaultSortedTxs

func (txs arr_tx) Len() int {
	return len(txs)
}
func (txs arr_tx) Less(i, j int) bool {
	return txs[i].GasPrice() > txs[j].GasPrice()
}
func (txs arr_tx) Swap(i, j int) {
	txs[i], txs[j] = txs[j], txs[i]
}

// these tx have a same gasPrice and addr.
type DefaultSortedTxs []*types.Transaction

func (d *DefaultSortedTxs) GasPrice() uint64 {
	tem := *d
	return tem[0].GasPrice
}
func (d *DefaultSortedTxs) Push(tx *types.Transaction) {
	tem := *d
	tem = append(tem, tx)
	*d = tem
}
func (d *DefaultSortedTxs) Pop() *types.Transaction {
	tem := *d
	if len(tem) == 0 {
		return nil
	}
	res := tem[0]
	if len(tem) < 2 {
		tem = make(DefaultSortedTxs, 0)
		*d = tem
	} else {
		*d = tem[1:]
	}
	return res
}
func (d *DefaultSortedTxs) Replace(tx *types.Transaction) {
	tem := *d
	for i, v := range tem {
		if v.Nonce == tx.Nonce {
			tem[i] = tx
		}
	}
	*d = tem
}
func (d *DefaultSortedTxs) Nonce() uint64 {
	tem := *d
	return tem[len(tem)-1].Nonce
}
func (d *DefaultSortedTxs) Address() types.Address {
	tem := *d
	return tem[0].From()
}

type QueenSortedTxs []*types.Transaction

func (txs QueenSortedTxs) Len() int {
	return len(txs)
}
func (txs QueenSortedTxs) Less(i, j int) bool {
	return txs[i].Nonce < txs[j].Nonce
}
func (txs QueenSortedTxs) Swap(i, j int) {
	txs[i], txs[j] = txs[j], txs[i]
}

type SubPool interface {
	Push(tx *types.Transaction)
	Pop() *types.Transaction
	GasPrice() uint64
	Nonce() uint64
	Address() types.Address
	Replace(tx *types.Transaction)
}

type PendingTxs []SubPool

func NewPendingTxs() PendingTxs {
	res := make([]SubPool, 0)
	return res
}
func (txs PendingTxs) Len() int {
	return len(txs)
}
func (txs PendingTxs) Less(i, j int) bool {
	if txs[i].Address() == txs[j].Address() {
		return txs[i].Nonce() < txs[j].Nonce()
	}
	return txs[i].GasPrice() > txs[j].GasPrice()
}
func (txs PendingTxs) Swap(i, j int) {
	txs[i], txs[j] = txs[j], txs[i]
}

type DefaultPool struct {
	StatDB  *statdb.IStatDB
	all     map[hash.Hash]bool
	txs     PendingTxs
	pending map[types.Address]arr_tx
	queue   map[types.Address]QueenSortedTxs
}

func (pool DefaultPool) Has() bool {
	return len(pool.txs) > 0
}
func (pool *DefaultPool) NewTx(tx *types.Transaction) {
	from := tx.From()
	account := pool.StatDB.Load(from)
	if account.Nonce >= tx.Nonce {
		return
	}

	nonce := account.Nonce
	pools, ok := pool.pending[from]
	if !ok {
		pools = make(arr_tx, 0)
		pool.pending[from] = pools
	}
	if len(pools) > 0 {
		last := pools[len(pools)-1]
		nonce = last.Nonce()
	}
	if tx.Nonce > nonce+1 {
		// 加到queue
		pool.addQueueTx(tx)
	} else if tx.Nonce == nonce+1 {
		// 加到pending，判断是否有queue的交易可以pop
		pool.addPendingTx(tx)
	} else {
		// replace
		pool.replacePendingTx(tx)
	}
}

func (pool *DefaultPool) addQueueTx(tx *types.Transaction) {
	txs := pool.queue[tx.From()]
	txs = append(txs, tx)
	// TODO 对txs进行排序
	sort.Sort(txs)
	pool.queue[tx.From()] = txs
}

// tx's nonce is max in pending+1, make it in the back of txs
func (pool *DefaultPool) addPendingTx(tx *types.Transaction) {
	from := tx.From()
	subpools := pool.pending[from]
	if len(subpools) == 0 {
		sub := make(DefaultSortedTxs, 0)
		sub = append(sub, tx)
		subpools = append(subpools, sub)
		pool.txs = append(pool.txs, &sub)
		pool.pending[from] = subpools
		sort.Sort(pool.txs)
	} else {
		last := subpools[len(subpools)-1]
		if last.GasPrice() <= tx.GasPrice {
			last = append(last, tx)
			pool.pending[from][len(subpools)-1] = last
			for i, v := range pool.txs {
				if v.Address() == from && v.Nonce() < tx.Nonce {
					pool.txs[i].Push(tx)
					break
				}
			}
			sort.Sort(pool.txs)
		} else {
			sub := make(DefaultSortedTxs, 0)
			sub = append(sub, tx)
			subpools = append(subpools, sub)
			pool.txs = append(pool.txs, &sub)
			pool.pending[from] = subpools
			sort.Sort(pool.txs)
		}
	}
	// TODO 更新queue中可以pop到pending中的交易
	quee := pool.queue[from]
	sort.Search(quee.Len(), func(i int) bool {
		if quee[i].Nonce == tx.Nonce+1 {
			pool.addPendingTx(quee[i])
		}
		return true
	})
}

func (pool *DefaultPool) replacePendingTx(tx *types.Transaction) {
	subpools := pool.pending[tx.From()]
	for _, sub := range subpools {
		if sub.Nonce() >= tx.Nonce {
			if tx.GasPrice >= sub.GasPrice() {
				sub.Replace(tx)
			}
			break
		}
	}
}

func (pool DefaultPool) Nonce(addr types.Address) uint64 {
	return 1
}

func (pool *DefaultPool) Pop() *types.Transaction {
	res := pool.txs[0].Pop()
	x := pool.txs[0].(*DefaultSortedTxs)
	if len(*x) == 0 {
		if len(pool.txs) == 1 {
			pool.txs = make(PendingTxs, 0)
		} else {
			pool.txs = pool.txs[1:]
		}
	}
	addr := res.From()
	nonce := res.Nonce
	for idx, v := range pool.pending[addr] {
		for i, v2 := range v {
			if v2.Nonce == nonce {
				if len(v) == 1 {
					if len(pool.pending[addr]) == 1 {
						delete(pool.pending, addr)

					} else {
						if idx == len(pool.pending[addr])-1 {
							pool.pending[addr] = pool.pending[addr][:idx]
						} else if idx == 0 {
							pool.pending[addr] = pool.pending[addr][1:]
						} else {
							pool.pending[addr] = append(pool.pending[addr][:idx], pool.pending[addr][idx+1:]...)

						}
					}
				} else {
					if i == 0 {
						pool.pending[addr][idx] = pool.pending[addr][idx][1:]
					} else if i == len(pool.pending[addr][idx])-1 {
						pool.pending[addr][idx] = pool.pending[addr][idx][:i]
					} else {
						pool.pending[addr][idx] = append(pool.pending[addr][idx][:i], pool.pending[addr][idx][i+1:]...)
					}
				}
				return res
			}
		}
	}
	return res
}

func (pool DefaultPool) SetStatRoot(root hash.Hash) {}

func (pool DefaultPool) NotifyTxEvent(txs []*types.Transaction) {}

func NewDefaultPool(db statdb.IStatDB) DefaultPool {
	return DefaultPool{
		&db,
		make(map[hash.Hash]bool),
		NewPendingTxs(),
		make(map[types.Address]arr_tx),
		make(map[types.Address]QueenSortedTxs),
	}
}
