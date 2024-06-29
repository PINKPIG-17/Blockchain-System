package statdb

import (
	"cxchain223/kvstore"
	"cxchain223/types"
	"cxchain223/utils/hash"
	"cxchain223/utils/rlp"
	"fmt"
)

type StatDB interface {
	SetStatRoot(root hash.Hash)
	Load(addr types.Address) *types.Account
	Store(addr types.Address, account types.Account)
}
type IStatDB struct {
	//Trie trie.ITrie
	Trie *kvstore.LevelDB
}

//	func NewIStatDB(db kvstore.KVDatabase, root hash.Hash) IStatDB {
//		return IStatDB{
//			trie.NewState(db,root),
//		}
//	}
func NewIStatDB(path string) IStatDB {
	return IStatDB{
		Trie: kvstore.NewLevelDB(path),
	}
}
func (i *IStatDB) SetStatRoot(root hash.Hash) {
	//e := i.Trie.(*trie.State)
	//e.

}
func (i *IStatDB) Load(addr types.Address) *types.Account {
	//data, err := i.Trie.Load(addr[:])
	data, err := i.Trie.Get(addr[:])
	if err != nil {
		fmt.Println("Bob addr:", addr)
		panic(err)
	}
	var account types.Account
	err = rlp.DecodeBytes(data, &account)
	if err != nil {
		panic(err)
	}
	return &account
}
func (i *IStatDB) Store(addr types.Address, account types.Account) {
	val, err := rlp.EncodeToBytes(account)
	if err != nil {
		panic(err)
	}
	//err = i.Trie.Store(addr[:], val)
	err = i.Trie.Put(addr[:], val)
	if err != nil {
		panic(err)
	}
}
