package statdb

import (
	"cxchain223/kvstore"
	"cxchain223/trie"
	"cxchain223/types"
	"cxchain223/utils/hash"
	"cxchain223/utils/rlp"
)

type StatDB interface {
	SetStatRoot(root hash.Hash)
	Load(addr types.Address) *types.Account
	Store(addr types.Address, account types.Account)
}

type statedbroot struct {
	state *trie.State
}

func (sta statedbroot) SetStatRoot(root hash.Hash) {
	var db kvstore.KVDatabase
	state := trie.NewState(db, root)
	sta.state = state
}

func (sta statedbroot) Load(addr types.Address) *types.Account {
	var byteaddr []byte
	copy(byteaddr, addr[:])
	accountValue, _ := sta.state.Load(byteaddr)
	var account types.Account
	rlp.DecodeBytes(accountValue, account)
	return &account
}

func (sta statedbroot) Store(addr types.Address, account types.Account) {
	accountValue, _ := rlp.EncodeToBytes(account)
	sta.state.Store(addr[:], accountValue)
}
