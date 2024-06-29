package types

import (
	"cxchain223/crypto"
	"cxchain223/crypto/secp256k1"
	"cxchain223/crypto/sha3"
	"cxchain223/utils/hash"
	"cxchain223/utils/hexutil"
	"cxchain223/utils/rlp"
	"errors"
	"fmt"
	"math/big"
)

type Receiption struct {
	TxHash  hash.Hash
	Status  int
	GasUsed uint64
}

type Transaction struct {
	txdata
	signature
}

type txdata struct {
	To       Address
	Nonce    uint64
	Value    uint64
	Gas      uint64
	GasPrice uint64
	Input    []byte
}

type signature struct {
	R, S *big.Int
	V    uint8
}

func (tx Transaction) From() Address {
	txdata := tx.txdata
	toSign, err := rlp.EncodeToBytes(txdata)
	fmt.Println(hexutil.Encode(toSign), err)
	msg := sha3.Keccak256(toSign)
	sig := make([]byte, 65)
	pubKey, err := secp256k1.RecoverPubkey(msg[:], sig)
	if err != nil {
		// TODO 返回一个空地址
		return [20]byte{}
	}
	return PubKeyToAddress(pubKey)
}

func (tx Transaction) VerifySignature() error {
	txdata := tx.txdata
	toSign, err := rlp.EncodeToBytes(txdata)
	fmt.Println(hexutil.Encode(toSign), err)
	msg := sha3.Keccak256(toSign)
	sig := make([]byte, 65)
	pubKey, err := secp256k1.RecoverPubkey(msg[:], sig)
	if err != nil {
		return errors.New("Invalid pubKey")
	}
	if !crypto.VerifySignature(pubKey, msg[:], tx.Bytes()) {
		return errors.New("Invalid signature")
	}
	return nil
}

func (tx Transaction) Bytes() []byte { // turn tx.signature to sign of byte type
	s := tx.signature
	rBytes := s.R.Bytes()
	sBytes := s.S.Bytes()
	vByte := []byte{s.V}

	if len(rBytes) < 32 {
		rBytes = append(make([]byte, 32-len(rBytes)), rBytes...)
	}
	if len(sBytes) < 32 {
		sBytes = append(make([]byte, 32-len(sBytes)), sBytes...)
	}

	sigBytes := append(rBytes, sBytes...)
	sigBytes = append(sigBytes, vByte...)

	return sigBytes
}
