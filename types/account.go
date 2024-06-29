package types

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"cxchain223/crypto/sha3"
	"cxchain223/utils/hash"
	"encoding/pem"
)

type Account struct {
	Amount     uint64
	Nonce      uint64
	CodeHash   hash.Hash
	PrivateKey []byte
}

func NewAccount(code []byte, amount uint64) Account {
	pk, err := rsa.GenerateKey(rand.Reader, 2048)
	d := x509.MarshalPKCS1PrivateKey(pk)
	block := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: d,
	}
	privkey := pem.EncodeToMemory(block)
	if err != nil {
		panic(err)
	}
	if code == nil {
		return Account{
			Amount:     amount,
			Nonce:      0,
			PrivateKey: privkey,
		}
	}
	return Account{
		Amount:     0,
		Nonce:      0,
		CodeHash:   sha3.Keccak256(code),
		PrivateKey: privkey,
	}
}
func parsePrivKey(pri []byte) *rsa.PrivateKey {
	blk, _ := pem.Decode(pri)
	privkey, _ := x509.ParsePKCS1PrivateKey(blk.Bytes)
	return privkey
}
func ParsePubKey(pub []byte) *rsa.PublicKey {
	blk, _ := pem.Decode(pub)
	p, _ := x509.ParsePKCS1PublicKey(blk.Bytes)
	return p
}
func (a Account) GetPrivKeyFromAccount() *rsa.PrivateKey {
	privkey := parsePrivKey(a.PrivateKey)
	return privkey
}
func (a Account) GetPubKeyFromAccount() *rsa.PublicKey {
	privkey := parsePrivKey(a.PrivateKey)
	pub := &privkey.PublicKey
	return pub
}
func (a Account) GetAddress() Address {
	privkey := parsePrivKey(a.PrivateKey)
	pub := &privkey.PublicKey
	d, _ := x509.MarshalPKIXPublicKey(pub)
	block := &pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: d,
	}
	pubkey := pem.EncodeToMemory(block)
	return PubKeyToAddress(pubkey)
}
