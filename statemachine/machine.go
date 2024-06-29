package statemachine

import (
	"goXChain/crypto/sha3"
	"goXChain/statdb"
	"goXChain/trie"
	"goXChain/types"
	"goXChain/utils/rlp"
	"errors"
	"fmt"
)

type IMachine interface {
	Execute(state trie.ITrie, tx types.Transaction) error
	Execute1(state statdb.StatDB, tx types.Transaction) *types.Receiption
}

type StateMachine struct {
}

func (m StateMachine) Execute1(state statdb.StatDB, tx types.Transaction) *types.Receiption {
	var receipt types.Receiption
	err := m.Execute(state.All(), tx)
	receipt.Status = 1
	if err != nil {
		receipt.Status = 0
		fmt.Println(err)
	}
	tx1, _ := rlp.EncodeToBytes(tx)
	receipt.TxHash = sha3.Keccak256(tx1)
	receipt.GasUsed = tx.Gas
	state.SetStatRoot(state.All().Root())
	return &receipt
}

func (m StateMachine) Execute(state trie.ITrie, tx types.Transaction) error {
	stat := state.Root()
	from := tx.Txdata.From
	to := tx.To
	value := tx.Value
	gasUsed := tx.Gas
	if tx.Gas < 2 {
		return errors.New("交易油费不足！")
	} else {
		gasUsed = 2
	}
	gasUsed = gasUsed * tx.GasPrice
	cost := value + gasUsed

	data, err := state.Load(from[:])
	if err != nil {
		return errors.New("账户查询失败！")
	}
	var account types.Account
	_ = rlp.DecodeBytes(data, &account)

	if account.Amount < cost {
		return errors.New("账户余额失败！")
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
	if stat == state.Root() {
		return errors.New("状态更新失败！")
	}
	return nil
}
