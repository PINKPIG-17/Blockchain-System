package types

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"cxchain223/utils/hash"
	"encoding/pem"
	"fmt"
	"math/big"
)

type Receiption struct {
	TxHash hash.Hash
	Status int
	// GasUsed int
	// Logs
}

func (r Receiption) ToString() string {
	return string(string(r.TxHash[:]) + string(r.Status) + "\n")
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
	from     []byte
}

func (t txdata) ToString() string {
	return string("To:" + string(t.To[:]) + "\nNonce:" + string(t.Nonce) +
		"\nValue:" + string(t.Value) + "\nGas:" + string(t.Gas) +
		"\nGasPrice:" + string(t.GasPrice) + "\nInput:" + string(t.Input) +
		"\nFrom:" + string(t.from) + "\n")
}

type signature struct {
	R, S *big.Int
	V    *big.Int
}

func NewTx(from *Account, to Address, nonce uint64, value uint64, gas uint64, gasPrice uint64, input []byte) *Transaction {
	pubkey, privkey := from.GetPubKeyFromAccount(), from.GetPrivKeyFromAccount()

	d, _ := x509.MarshalPKIXPublicKey(pubkey)
	block := &pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: d,
	}
	pub := pem.EncodeToMemory(block)
	sum := sha256.Sum256(input)
	sig, err := rsa.SignPKCS1v15(rand.Reader, privkey, crypto.SHA256, sum[:])
	if err != nil {
		fmt.Println("ERROR TO SIGN")
		return nil
	}
	r := new(big.Int).SetBytes(sig[:128])
	s := new(big.Int).SetBytes(sig[128:255])
	v := new(big.Int).SetBytes([]byte{sig[255] + 27})
	return &Transaction{
		txdata{
			to,
			nonce,
			value,
			gas,
			gasPrice,
			input,
			pub,
		},
		signature{
			R: r,
			S: s,
			V: v,
		},
	}
}

func (tx Transaction) From() Address {
	return PubKeyToAddress(tx.from)
}
func (tx Transaction) Verify() bool {
	r, s, v := tx.R, tx.S, tx.V
	d := v.Bytes()
	bt := d[0] - 27
	sig := append(r.Bytes(), s.Bytes()...)
	sig = append(sig, bt)
	sum := sha256.Sum256(tx.Input)
	pub := ParsePubKey(tx.from)
	err := rsa.VerifyPKCS1v15(pub, crypto.SHA256, sum[:], sig)
	if err != nil {
		return false
	}
	return true
}
