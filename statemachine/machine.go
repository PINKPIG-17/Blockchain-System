package statemachine

import (
	"cxchain223/crypto/sha3"
	"cxchain223/statdb"
	"cxchain223/trie"
	"cxchain223/types"
	"cxchain223/utils/rlp"
)

type IMachine interface {
	Execute(state trie.ITrie, tx types.Transaction)
	Execute1(state statdb.StatDB, tx types.Transaction) *types.Receiption
}

type StateMachine struct {
}

func (m StateMachine) Execute(state trie.ITrie, tx types.Transaction) {
	from := tx.From()
	to := tx.To
	value := tx.Value
	gasUsed := tx.Gas
	if tx.Gas < 21000 {
		return
	} else {
		gasUsed = 21000
	}
	gasUsed = gasUsed * tx.GasPrice
	cost := value + gasUsed

	data, err := state.Load(from[:])
	if err != nil {
		return
	}
	var account types.Account
	_ = rlp.DecodeBytes(data, &account)

	if account.Amount < cost {
		return
	}

	account.Amount = account.Amount - cost
	data, err = rlp.EncodeToBytes(account)

	state.Store(from[:], data)

	data, err = state.Load(to[:])
	var toAccount types.Account
	if err != nil {
		toAccount = types.Account{}
	} else {
		rlp.DecodeBytes(data, &toAccount)
	}
	toAccount.Amount = toAccount.Amount + value
	data, err = rlp.EncodeToBytes(toAccount)

	state.Store(to[:], data)
}

func (m StateMachine) Execute1(state statdb.StatDB, tx types.Transaction) *types.Receiption {
	//TODO implement me
	from := tx.From()
	to := tx.To
	value := tx.Value
	gasUsed := tx.Gas
	if tx.Gas < 21000 {
		return nil
	} else {
		gasUsed = 21000
	}
	gasUsed = gasUsed * tx.GasPrice //total gas
	cost := value + gasUsed
	account := state.Load(from) //the sender account
	if account == nil {
		return nil
	}
	if account.Amount < gasUsed {
		return nil
	}
	account.Amount -= cost
	state.Store(from, *account)
	_to := state.Load(to) //to_account
	if _to == nil {
		_to = &types.Account{} //new an to_account
	}
	_to.Amount += value // transfer
	state.Store(to, *_to)
	toSign, err := rlp.EncodeToBytes(tx) //encode
	if err != nil {
		return &types.Receiption{
			Status: 0,
		}
	}
	txHash := sha3.Keccak256(toSign) //sign
	receipt := &types.Receiption{
		TxHash:  txHash,
		Status:  1,
		GasUsed: gasUsed,
	}
	return receipt
}
