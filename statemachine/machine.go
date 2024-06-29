package statemachine

import (
	"cxchain223/statdb"
	"cxchain223/trie"
	"cxchain223/types"
	"cxchain223/utils/hash"
	"cxchain223/utils/rlp"
	"fmt"
)

type IMachine interface {
	Execute(state trie.ITrie, tx types.Transaction)
	Execute1(state statdb.StatDB, tx types.Transaction) *types.Receiption
}

type StateMachine struct {
}

func NewStateMachine() StateMachine {
	return StateMachine{}
}
func (m *StateMachine) Execute(state trie.ITrie, tx types.Transaction) {
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
func (m *StateMachine) Execute1(state statdb.StatDB, tx types.Transaction) *types.Receiption {
	if !tx.Verify() {
		return nil
	}
	data, err := rlp.EncodeToBytes(tx)
	if err != nil {
		panic(err)
	}
	h := hash.BytesToHash(data)
	receipt := types.Receiption{
		h,
		500,
	}
	from := tx.From()
	to := tx.To
	value := tx.Value
	gasUsed := tx.Gas
	if tx.Gas < 21000 {
		return &receipt
	} else {
		gasUsed = 21000
	}
	gasUsed = gasUsed * tx.GasPrice
	cost := value + gasUsed
	account := state.Load(from)
	if account.Amount < cost {
		return &receipt
	}
	account.Amount = account.Amount - cost
	state.Store(from, *account)

	account = state.Load(to)
	account.Amount = account.Amount + value
	state.Store(to, *account)
	receipt = types.Receiption{
		h,
		200,
	}
	fmt.Printf("Excute TX: %v to %v with %d eth.\n", from[:], tx.To[:], tx.Value)
	return &receipt
}
