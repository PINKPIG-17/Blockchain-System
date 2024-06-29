package types 
import (
	"goXChain/crypto/sha3"
)

type Address [20]byte

func PubKeyToAddress(pub []byte) Address {
	h := sha3.Keccak256(pub)
	var addr Address
	copy(addr[:],h[:20])
	return addr
}

func AddressFromBytes(data []byte) *Address {
	var a [20]byte
	copy(a[:], data)
	return (*Address)(&a)
}
