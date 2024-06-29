package blockchain

import (
	"cxchain223/crypto/sha3"
	"cxchain223/trie"
	"cxchain223/txpool"
	"cxchain223/types"
	"cxchain223/utils/hash"
	"cxchain223/utils/rlp"
	"strconv"
	"time"
)

type Block struct {
	Header
	Body
}
type Header struct {
	Root       hash.Hash
	ParentHash hash.Hash
	Height     uint64
	Coinbase   types.Address
	Timestamp  uint64

	Nonce uint64
}

var EmptyHeader = Header{
	Root:       sha3.Keccak256([]byte("block_")),
	ParentHash: hash.EmptyHash,
	Height:     0,
	Timestamp:  uint64(time.Now().Unix()),
}

type Body struct {
	Transactions []types.Transaction
	Receiptions  []types.Receiption
}

func (header Header) Hash() hash.Hash {
	data, _ := rlp.EncodeToBytes(header)
	return sha3.Keccak256(data)
}

func NewHeader(parent *Header) *Header {
	if parent.ParentHash == hash.EmptyHash {
		return &EmptyHeader
	}
	i := int(parent.Height + 1)
	str := "block_" + strconv.Itoa(i)
	return &Header{
		Root:       sha3.Keccak256([]byte(str)),
		ParentHash: parent.Hash(),
		Height:     parent.Height + 1,
		Timestamp:  uint64(time.Now().Unix()),
	}
}

func NewBlock() *Body {
	return &Body{
		Transactions: make([]types.Transaction, 0),
		Receiptions:  make([]types.Receiption, 0),
	}
}

type Blockchain struct {
	CurrentHeader *Header
	Db            trie.ITrie
	Txpool        txpool.TxPool
}
