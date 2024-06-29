package main

import (
	"cxchain223/blockchain"
	"cxchain223/kvstore"
	maker2 "cxchain223/maker"
	"cxchain223/statdb"
	"cxchain223/statemachine"
	"cxchain223/trie"
	"cxchain223/txpool"
	"cxchain223/types"
)

func main() {
	//db := statdb.NewIStatDB(kvstore.NewLevelDB("./leveldb"), trie.EmptyHash)
	Stadb := statdb.NewIStatDB("./Staleveldb")
	datedb := trie.NewState(kvstore.NewLevelDB("./leveldb"), trie.EmptyHash)
	txpol := txpool.NewDefaultPool(Stadb)
	machine := statemachine.NewStateMachine()
	rootHeader := blockchain.EmptyHeader
	chain := blockchain.Blockchain{
		&rootHeader,
		//db.Trie,
		datedb,
		&txpol,
	}

	maker := maker2.NewBlockMaker(&txpol, &Stadb, &machine, &chain)

	alice := types.NewAccount(nil, 10000000)
	bob := types.NewAccount(nil, 10000000)
	addr_alice := alice.GetAddress()
	addr_bob := bob.GetAddress()
	Stadb.Store(addr_alice, alice)
	Stadb.Store(addr_bob, bob)
	for i := 0; i < 10; i++ {
		tx := types.NewTx(&alice, addr_bob, uint64(i+1), 1, 210000, 1, []byte("Send 1 eth."))
		txpol.NewTx(tx)
	}
	for i := 0; i < 10; i++ {
		tx := types.NewTx(&bob, addr_alice, uint64(i+1), 2, 200000, 1, []byte("Send 2 eth."))
		txpol.NewTx(tx)
	}
	maker.Seal()
}

//	go GenerateTX(alice, addr_bob, txpol)
//	maker.SealWithTime(3 * time.Minute)
//	end := time.After(70 * time.Second)
//	for {
//
//		select {
//		case <-end:
//			{
//				return
//			}
//		}
//	}
//}
//func GenerateTX(alice types.Account, addr_bob types.Address, txpol txpool.TxPool) {
//	t := time.NewTicker(time.Second)
//	end := time.After(time.Minute)
//	for {
//		select {
//		case <-t.C:
//			tx := types.NewTx(&alice, addr_bob, alice.Nonce+1, 1, 10, 1, []byte("Send 1 eth."))
//			txpol.NewTx(tx)
//		case <-end:
//			return
//		}
//	}
//}
