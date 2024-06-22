package types

import (
	"cxchain223/crypto/sha3"
)

type Address [20]byte

func PubKeyToAddress(pub []byte) Address {
	h := sha3.Keccak256(pub)
	var addr Address
	copy(addr[:], h[len(h)-20:])
	return addr
}
