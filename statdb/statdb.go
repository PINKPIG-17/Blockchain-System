package statdb

import (
	"goXChain/trie"
	"goXChain/types"
	"goXChain/utils/hash"
)

type StatDB interface {
	All() *trie.State
	SetStatRoot(root hash.Hash)
	Load(addr types.Address) *types.Account
	Store(addr types.Address, account types.Account)
}

type StatDb struct {
	root hash.Hash
	db   trie.State
}

func NewStatDb(db trie.State) *StatDb {
	var sta StatDb
	sta.root = db.Root()
	sta.db = db
	return &sta
}

func (statDb *StatDb) All() *trie.State {
	return &statDb.db
}

func (statDb *StatDb) SetStatRoot(root hash.Hash) {
	statDb.root = root
}

func (statDb *StatDb) Load(addr types.Address) *types.Account {
	account, err := statDb.db.Load(addr[:])
	if err != nil {
		return nil
	}
	return types.AccountFromBytes(account)
}

func (statDb *StatDb) Store(addr types.Address, account types.Account) {
	bytes := addr[:]
	err := statDb.db.Store(bytes, account.Bytes())
	if err != nil {
		return
	}
}