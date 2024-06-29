package types

import (
	"goXChain/utils/hash"
	"goXChain/utils/rlp"
)

type Account struct {
	Amount   uint64
	Nonce    uint64
	CodeHash hash.Hash
	Root     hash.Hash
}

func (account Account) Bytes() []byte {
	data, _ := rlp.EncodeToBytes(account)
	return data
}

func AccountFromBytes(data []byte) *Account {
	var account Account

	_ = rlp.DecodeBytes(data, &account)

	return &account
}
