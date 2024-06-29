package blockchain

import (
	"cxchain223/crypto/sha3"
	"cxchain223/trie"
	"cxchain223/txpool"
	"cxchain223/types"
	"cxchain223/utils/hash"
	"cxchain223/utils/rlp"
	"strconv"
	"strings"
	"time"
)

type ChainConfig struct {
	Duration   time.Duration
	Coinbase   types.Address
	Difficulty uint64
}

var DefaultConfig = ChainConfig{
	Difficulty: 0,
	Duration:   time.Second,
	Coinbase:   [20]byte{},
}

type Block struct {
	Header
	Body
}

func (b Block) ToString() string {
	return b.Header.ToString() + b.Body.ToString()
}

type Header struct {
	Root       hash.Hash
	ParentHash hash.Hash
	Height     uint64
	Coinbase   types.Address
	Timestamp  uint64
	Nonce      uint64
}

func (header *Header) ToString() string {
	return string("Root:" + header.Root.String() + "\nParentHash:" + header.ParentHash.String() +
		"\nHeight:" + string(header.Height) + "\nCoinbase:" + string(header.Coinbase[:]) +
		"\nTimestamp:" + string(header.Timestamp) + "\nNonce:" + string(header.Nonce) + "\n")
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

func (b *Body) ToString() string {
	s := "Trans:\n"
	for _, v := range b.Transactions {
		s += v.ToString()
	}
	s += "\nReceipt:\n"
	for _, v := range b.Receiptions {
		s += v.ToString()
	}
	return s
}
func (header Header) Hash() hash.Hash {
	data, _ := rlp.EncodeToBytes(header)
	return sha3.Keccak256(data)
}
func mine(nonce uint64) uint64 {
	n := int(nonce)
	matcher := ""
	for i := 0; i < int(DefaultConfig.Difficulty); i++ {
		matcher += "0"
	}
	for i := n; ; i++ {
		h := sha3.Keccak256([]byte{byte(n)})
		if strings.HasSuffix(h.String(), matcher) {
			return uint64(i)
		}
	}
}
func NewHeader(parent *Header) *Header {
	if parent.ParentHash == hash.EmptyHash {
		i := int(parent.Height + 1)
		str := "block_" + strconv.Itoa(i)
		start := time.Now()
		nonce := mine(1)
		end := time.Now()
		m := end.Sub(start).Minutes()
		if m > 5 {
			DefaultConfig.Difficulty++
		} else if m < 3 {
			DefaultConfig.Difficulty--
		}
		res := EmptyHeader
		res.ParentHash = parent.ParentHash
		res.Nonce = nonce
		res.Root = sha3.Keccak256([]byte(str))
		res.Height = parent.Height + 1
		return &res
	}
	start := time.Now()
	nonce := mine(parent.Nonce + 1)
	end := time.Now()
	m := end.Sub(start).Minutes()
	if m > 5 {
		DefaultConfig.Difficulty++
	} else if m < 3 {
		DefaultConfig.Difficulty--
	}
	i := int(parent.Height + 1)
	str := "block_" + strconv.Itoa(i)
	return &Header{
		Root:       sha3.Keccak256([]byte(str)),
		ParentHash: parent.Hash(),
		Height:     parent.Height + 1,
		Timestamp:  uint64(time.Now().Unix()),
		Nonce:      nonce,
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
