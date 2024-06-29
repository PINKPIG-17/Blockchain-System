package blockchain

import (
	"goXChain/crypto/sha3"
	"goXChain/statdb"
	"goXChain/trie"
	"goXChain/txpool"
	"goXChain/types"
	"goXChain/utils/hash"
	"goXChain/utils/rlp"
)

type Header struct {
	Root       hash.Hash
	ParentHash hash.Hash
	Height     uint64
	Coinbase   types.Address
	Timestamp  uint64

	Nonce uint64
}

func (header Header) Hash() hash.Hash {
	data, _ := rlp.EncodeToBytes(header)
	return sha3.Keccak256(data)
}

type Body struct {
	Transactions []types.Transaction
	Receiptions  []types.Receiption
}

func NewHeader(parent Header, statdb statdb.StatDB) *Header {
	return &Header{
		Root:       statdb.All().Root(),
		ParentHash: parent.Hash(),
		Height:     parent.Height + 1,
	}
}

func NewBody() *Body {
	return &Body{
		Transactions: make([]types.Transaction, 0),
		Receiptions:  make([]types.Receiption, 0),
	}
}

type Blockchain struct {
	CurrentHeader Header
	State         trie.ITrie
	Txpool        txpool.TxPool
}

func (b *Blockchain) setCurrentHeader(header Header) {
	b.CurrentHeader = header
}

func (b *Blockchain) setState(s trie.ITrie) {
	b.State = s
}

func (b *Blockchain) setTxpool(tx txpool.TxPool) {
	b.Txpool = tx
}
