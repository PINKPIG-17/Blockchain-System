package types

import (
	"cxchain223/crypto/sha3"
)

type Address [20]byte

const AddressLength = 20
const HashLen = 32

func PubKeyToAddress(pub []byte) Address {
	h := sha3.Keccak256(pub).Bytes()
	addr := make([]byte, 20)
	// TODO hash得到addr
	if len(h) > len(addr) {
		h = h[len(h)-HashLen:]
	} else {
		copy(addr[HashLen-len(h):], h)
	}
	res := *new([20]byte)
	for i, _ := range res {
		res[i] = h[i]
	}
	return res
}
