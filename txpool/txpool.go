package txpool

import (
	"crypto/ecdsa"
	"goXChain/types"
	"goXChain/utils/hash"
)

type TxPool interface {
	SetStatRoot(root hash.Hash)
	NewTx(tx *types.Transaction, pubk ecdsa.PublicKey)
	Pop() *types.Transaction
	NotifyTxEvent(txs []*types.Transaction, pubk []ecdsa.PublicKey)
}
